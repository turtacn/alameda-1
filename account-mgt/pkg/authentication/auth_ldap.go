package authentication

import (
	//"errors"
	"fmt"
	"gopkg.in/ldap.v2"
	"strings"
	//Logger "github.com/prophetstor-ai/common-lib/pkg/log"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	"github.com/rs/xid"
	"sort"
	"strconv"
	"time"
)

var LdapIP = "127.0.0.1"
var LdapPort = uint16(389)

var LdapBaseDN = "dc=prophetstor,dc=com"
var LdapUserAttributes = []string{
	"ou",
	"uid",
	"sn",
	"employeeType",
	"mail",
	"description",
	"userPassword",
}

const (
	LdapAdminID = "admin"
	LdapAdminPW = "password"
)

type AuthLdap struct {
	Address string
}

func (c *AuthLdap) Authenticate(authUserInfo *AuthUserInfo) (string, error) {
	scope.Info("authentication.Authenticate: starting...")

	lconn, err := c.loginLDAP(authUserInfo)
	if lconn != nil {
		defer lconn.Close()
	}

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.Authenticate: %v", err.Error()))
		return "", err
	}

	searchRequest := ldap.NewSearchRequest(
		authUserInfo.DN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		fmt.Sprintf("(uid=%s)", authUserInfo.Name),
		LdapUserAttributes,
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.Authenticate: %v", err.Error()))
		return "", Errors.NewError(Errors.ReasonUserNotFound, authUserInfo.Name)
	}
	c.loadUserAllData(authUserInfo, sr.Entries[0])

	claims := make(map[string]string)
	claims["name"] = authUserInfo.Name
	claims["namespace"] = authUserInfo.Namespace
	claims["domainName"] = authUserInfo.DomainName
	claims["role"] = authUserInfo.Role

	jwtstr := JwtGenerate(claims, TokenDuration)
	if jwtstr == "" {
		scope.Error(fmt.Sprintf("authentication.Authenticate: fail to generate JWT"))
		return "", Errors.NewError(Errors.ReasonFailedToGenJWT)
	}

	scope.Info(fmt.Sprintf("authentication.Authenticate: authenticate user(%s) successfully.", authUserInfo.Name))
	return jwtstr, nil
}

func (c *AuthLdap) Validate(authUserInfo *AuthUserInfo, token string) (string, error) {
	return Validate(authUserInfo, token)
}

func (c *AuthLdap) ChangePassword(authUserInfo *AuthUserInfo, newPassword string) error {
	scope.Info("authentication.ChangePassword: starting...")
	lconn, err := c.loginLDAP(authUserInfo)
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ChangePassword: %v", err.Error()))
		return err
	}

	passwordModifyRequest := ldap.NewPasswordModifyRequest("", authUserInfo.Password, newPassword)
	_, err = lconn.PasswordModify(passwordModifyRequest)

	if err != nil {
		scope.Error(fmt.Sprintf("Password could not be changed: %v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	authUserInfo.Password = newPassword

	scope.Info(fmt.Sprintf("authentication.ChangePassword: change user(%s) password successfully", authUserInfo.Name))
	return nil
}

func (c *AuthLdap) CreateUser(authUserInfo *AuthUserInfo, joinDomain bool) error {
	scope.Info("authentication.CreateUser: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.CreateUser: %v", err.Error()))
		return err
	}

	//check user exist
	userExist, err := c.IsUserExist(authUserInfo.Name)
	if err != nil {
		scope.Errorf("authentication.CreateUser failed: %s", err.Error())
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	if userExist {
		scope.Errorf("authentication.CreateUser failed: User(%s) already exist", authUserInfo.Name)
		return Errors.NewError(Errors.ReasonUserAlreadyExist, authUserInfo.Name)
	}

	authUserInfo.SendConfirmCount = 0
	authUserInfo.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	authUserInfo.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	authUserInfo.Timezone = time.Now().UTC().Format("-0700")
	authUserInfo.Namespace = c.selectNamespace()

	attrs := c.genDescriptionAttrs(authUserInfo)

	//-----------------------------------------------------------
	if joinDomain == false {
		newOu, _ := c.GenNewOUName(authUserInfo.DomainName)
		authUserInfo.DomainName = newOu

		authUserInfo.DN = fmt.Sprintf("uid=%s,ou=%s,%s", authUserInfo.Name, authUserInfo.DomainName, LdapBaseDN)
	}

	//create ou--------------------------------------------------
	ouAttrs := []string{}
	ouRequest := ldap.NewAddRequest(fmt.Sprintf("ou=%s,%s", authUserInfo.DomainName, LdapBaseDN))
	ouRequest.Attribute("objectClass", []string{"top", "organizationalUnit"})
	ouRequest.Attribute("ou", []string{authUserInfo.DomainName})
	ouAttrs = append(ouAttrs, fmt.Sprintf("namespace=%s", authUserInfo.Namespace))
	ouRequest.Attribute("description", ouAttrs)

	err = lconn.Add(ouRequest)
	if err != nil {
		errMsg := fmt.Sprintf("%v", err.Error())
		if strings.Contains(errMsg, "Entry Already Exists") == false {
			scope.Error(fmt.Sprintf("CreateOu failed: %v", err.Error()))
			return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
		}
	}

	//create user--------------------------------------------------
	userRequest := ldap.NewAddRequest(authUserInfo.DN)
	sshaPassword, err := passwordSshaEncode(authUserInfo.Password)
	if err != nil {
		scope.Error(fmt.Sprintf("%v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	authUserInfo.ID = xid.New().String()
	//userRequest.Attribute("objectClass", []string{"account", "simpleSecurityObject","top"})
	userRequest.Attribute("objectClass", []string{"inetOrgPerson", "top"})
	userRequest.Attribute("ou", []string{authUserInfo.DomainName})
	userRequest.Attribute("cn", []string{authUserInfo.Name})
	userRequest.Attribute("sn", []string{authUserInfo.ID})
	userRequest.Attribute("uid", []string{authUserInfo.Name})
	userRequest.Attribute("userPassword", []string{sshaPassword})
	userRequest.Attribute("mail", []string{authUserInfo.Email})
	userRequest.Attribute("employeeType", []string{authUserInfo.Role})

	userRequest.Attribute("description", attrs)

	err = lconn.Add(userRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("%v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	scope.Info(fmt.Sprintf("authentication.CreateUser: create user(%s) successfully", authUserInfo.Name))
	return nil
}

func (c *AuthLdap) CreateUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo, joinDomain bool) error {
	return c.CreateUser(ownerInfo, joinDomain)
}

func (c *AuthLdap) selectNamespace() string {
	return "federator-ai"
}

func (c *AuthLdap) ReadUser(authUserInfo *AuthUserInfo) error {
	scope.Info("authentication.ReadUser: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadUser: %v", err.Error()))
		return err
	}

	if authUserInfo.DomainName == "" {
		ou, err := c.SearchOUByUser(authUserInfo.Name)
		if err != nil {
			scope.Error(fmt.Sprintf("authentication.ReadUser: %v", err.Error()))
			return err
		}

		authUserInfo.DN = NewAuthUserInfo(ou, authUserInfo.Name).DN
		authUserInfo.DomainName = ou
	}

	searchRequest := ldap.NewSearchRequest(
		authUserInfo.DN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		//fmt.Sprintf("(&(uid:dn:=%s)(ou:dn:=%s))", user.Name, user.Ou),
		//fmt.Sprintf("(&(uid=%s)(ou=%s))", authUserInfo.Name, authUserInfo.Domain),
		fmt.Sprintf("(uid=%s)", authUserInfo.Name),
		LdapUserAttributes,
		nil)

	sr, err := lconn.Search(searchRequest)

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadUser: %v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		scope.Error(fmt.Sprintf("user %s is not found", authUserInfo.Name))
		return Errors.NewError(Errors.ReasonUserNotFound, authUserInfo.Name)
	}

	c.loadUserAllData(authUserInfo, sr.Entries[0])
	scope.Info(fmt.Sprintf("authentication.ReadUser: read user(%s) successfully.", authUserInfo.Name))

	return nil
}

func (c *AuthLdap) ReadUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error {
	return c.ReadUser(ownerInfo)
}

func (c *AuthLdap) GetUserListByDomain(authUserInfoList *[]AuthUserInfo, ou string, limit int, page int) error {
	scope.Info("authentication.GetUserListByDomain: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadUserList: %v", err.Error()))
		return err
	}

	tempBaseDN := ""
	tempFilter := ""

	if ou == "" {
		tempBaseDN = LdapBaseDN
		tempFilter = "(&(objectClass=inetOrgPerson)(!uid=_adm))"
	} else {
		tempBaseDN = fmt.Sprintf("ou=%s,%s", ou, LdapBaseDN)
		tempFilter = fmt.Sprintf("(&(objectClass=inetOrgPerson))")
	}

	paging := ldap.NewControlPaging(uint32(limit))
	searchRequest := ldap.NewSearchRequest(
		tempBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		tempFilter,
		LdapUserAttributes,
		[]ldap.Control{paging})

	sr, err := lconn.Search(searchRequest)

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetUserListByDomain: %v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	}

	for i := 0; i < page; i++ {
		pagingResult := ldap.FindControl(sr.Controls, ldap.ControlTypePaging)
		cookie := pagingResult.(*ldap.ControlPaging).Cookie
		paging.SetCookie(cookie)
		sr, err = lconn.Search(searchRequest)
	}

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetUserListByDomain: %v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		scope.Error(fmt.Sprintf("ou %s is not found", ou))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	}

	for _, entry := range sr.Entries {
		tempUserInfo := AuthUserInfo{}
		c.loadUserAllData(&tempUserInfo, entry)
		*authUserInfoList = append(*authUserInfoList, tempUserInfo)
	}

	scope.Info(fmt.Sprintf("authentication.GetUserListByDomain: read user by Ou successfully."))

	return nil
}

func (c *AuthLdap) GetUserListByDomainWithCaller(callerInfo *AuthUserInfo, ownerInfoList *[]AuthUserInfo, domain string, limit int, page int) error {
	return c.GetUserListByDomain(ownerInfoList, domain, limit, page)
}

func (c *AuthLdap) UpdateUser(authUserInfo *AuthUserInfo) error {
	scope.Info("authentication.UpdateUser starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.UpdateUser: %v", err.Error()))
		return err
	}

	authUserInfo.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	attrs := c.genDescriptionAttrs(authUserInfo)

	attrModifyRequest := ldap.NewModifyRequest(authUserInfo.DN)
	attrModifyRequest.ReplaceAttributes = []ldap.PartialAttribute{
		{
			Type: "description",
			Vals: attrs,
		},
	}

	//other attr
	role := ldap.PartialAttribute{
		Type: "employeeType",
		Vals: []string{authUserInfo.Role},
	}
	attrModifyRequest.ReplaceAttributes = append(attrModifyRequest.ReplaceAttributes, role)

	email := ldap.PartialAttribute{
		Type: "mail",
		Vals: []string{authUserInfo.Email},
	}
	attrModifyRequest.ReplaceAttributes = append(attrModifyRequest.ReplaceAttributes, email)

	if strings.Index(authUserInfo.Password, "{SSHA}") != 0 {
		authUserInfo.Password, err = passwordSshaEncode(authUserInfo.Password)
		if err != nil {
			scope.Error(fmt.Sprintf("%v", err.Error()))
			return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
		}
	}

	password := ldap.PartialAttribute{
		Type: "userPassword",
		Vals: []string{authUserInfo.Password},
	}
	attrModifyRequest.ReplaceAttributes = append(attrModifyRequest.ReplaceAttributes, password)

	//update
	err = lconn.Modify(attrModifyRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("%v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	//get user new data
	c.ReadUser(authUserInfo)

	scope.Info(fmt.Sprintf("authentication.UpdateUser: update user(%s) successfully.", authUserInfo.Name))
	return nil
}

func (c *AuthLdap) UpdateUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error {
	return c.UpdateUser(ownerInfo)
}

func (c *AuthLdap) DeleteUser(ownerInfo *AuthUserInfo) error {
	return c.deleteUser(ownerInfo, true)
}

func (c *AuthLdap) DeleteUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error {
	if callerInfo.Role == "super" {
		return c.deleteUser(ownerInfo, false)
	} else {
		return c.deleteUser(ownerInfo, true)
	}
}

func (c *AuthLdap) deleteUser(authUserInfo *AuthUserInfo, checkLastAdmin bool) error {
	scope.Info("authentication.DeleteUser: starting...")

	ouSubEntryCount := -1
	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.DeleteUser: %v", err.Error()))
		return err
	}

	//step2. check admin count
	if authUserInfo.Role == "domain-admin" && checkLastAdmin {
		tempUserList := make([]AuthUserInfo, 0)

		err = c.GetUserListByDomain(&tempUserList, authUserInfo.DomainName, 1000000, 0)
		if err != nil {
			scope.Error(fmt.Sprintf("authentication.DeleteUser: %v", err.Error()))
			return err
		}

		adminCount := 0
		for _, user := range tempUserList {
			if user.Role == "domain-admin" {
				adminCount++
			}
		}

		if adminCount <= 1 {
			scope.Error(fmt.Sprintf("authentication.DeleteUser: only one admin in domain(%s)", authUserInfo.DomainName))
			return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
		}
	}

	//step3. delete user
	userDel := ldap.NewDelRequest(authUserInfo.DN, nil)
	err = lconn.Del(userDel)

	if err != nil {
		scope.Error(fmt.Sprintf("%v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
	}

	//delete ou
	ouSubEntryCount, err = c.readOuSubEntryCount(authUserInfo.DomainName)
	if ouSubEntryCount == 0 {
		ouDel := ldap.NewDelRequest(fmt.Sprintf("ou=%s,%s", authUserInfo.DomainName, LdapBaseDN), nil)
		err = lconn.Del(ouDel)
		if err != nil {
			scope.Error(fmt.Sprintf("%v", err.Error()))
			return Errors.NewError(Errors.ReasonFailedToUpdateDB, "ldap")
		}
	}

	scope.Info(fmt.Sprintf("authentication.DeleteUser: deleted user(%s) successfully.", authUserInfo.Name))
	return nil
}

func (c *AuthLdap) loginLDAP(authUserInfo *AuthUserInfo) (*ldap.Conn, error) {
	address := ""
	if c.Address != "" {
		address = c.Address
	} else {
		address = fmt.Sprintf("%s:%d", LdapIP, LdapPort)
	}
	lconn, err := ldap.Dial("tcp", address)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.userLogin: %v", err.Error()))
		return lconn, Errors.NewError(Errors.ReasonFailedToConnectDB, "ldap")
	}

	//only username -> use username to find domain
	if authUserInfo.DomainName == "" {
		ou, err := c.SearchOUByUser(authUserInfo.Name)

		if err != nil {
			scope.Error(fmt.Sprintf("authentication.userLogin: %v", err.Error()))
			//return lconn, err
		}

		authUserInfo.DomainName = ou
		authUserInfo.DN = fmt.Sprintf("uid=%s,ou=%s,%s", authUserInfo.Name, authUserInfo.DomainName, LdapBaseDN)
	}

	err = lconn.Bind(authUserInfo.DN, authUserInfo.Password)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.userLogin: %v", err.Error()))
		return lconn, Errors.NewError(Errors.ReasonInvalidCredential)
	}

	return lconn, err
}

func (c *AuthLdap) adminLoginLDAP() (*ldap.Conn, error) {
	address := ""
	if c.Address != "" {
		address = c.Address
	} else {
		address = fmt.Sprintf("%s:%d", LdapIP, LdapPort)
	}
	lconn, err := ldap.Dial("tcp", address)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.userLogin: %v", err.Error()))
		return lconn, Errors.NewError(Errors.ReasonFailedToConnectDB, "ldap")
	}

	dn := fmt.Sprintf("cn=%s,%s", LdapAdminID, LdapBaseDN)
	err = lconn.Bind(dn, LdapAdminPW)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.userLogin: %v", err.Error()))
		return lconn, Errors.NewError(Errors.ReasonInvalidCredential)
	}

	return lconn, err
}

func (c *AuthLdap) readOuSubEntryCount(ouName string) (int, error) {
	scope.Info("authentication.ReadOuSubEntryCount: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadOuSubEntryCount: %v", err.Error()))
		return -1, err
	}

	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("ou=%s,%s", ouName, LdapBaseDN),
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		[]string{},
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadOuSubEntryCount: %v", err.Error()))
		return -1, Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		scope.Error(fmt.Sprintf("ou %s is not found", ouName))
		return -1, Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	}

	return len(sr.Entries), nil
}

func (c *AuthLdap) loadUserAllData(authUserInfo *AuthUserInfo, entry *ldap.Entry) {
	uid := entry.GetAttributeValues("uid")
	for _, attr := range uid {
		authUserInfo.Name = attr
	}

	ou := entry.GetAttributeValues("ou")
	for _, attr := range ou {
		authUserInfo.DomainName = attr
	}

	employeeType := entry.GetAttributeValues("employeeType")
	for _, attr := range employeeType {
		authUserInfo.Role = attr
	}

	mail := entry.GetAttributeValues("mail")
	for _, attr := range mail {
		authUserInfo.Email = attr
	}

	password := entry.GetAttributeValues("userPassword")
	for _, attr := range password {
		authUserInfo.Password = attr
	}

	sn := entry.GetAttributeValues("sn")
	for _, attr := range sn {
		authUserInfo.ID = attr
	}

	c.loadUserDescriptionData(authUserInfo, entry)
}

func (c *AuthLdap) loadUserDescriptionData(authUserInfo *AuthUserInfo, entry *ldap.Entry) {
	description := entry.GetAttributeValues("description")
	attrMap := map[string]string{}
	clusters := []ClusterInfo{}
	for _, attr := range description {
		attr = strings.TrimSpace(attr)
		i := strings.Index(attr, "=")
		if i > 0 {
			key := strings.TrimSpace(attr[0:i])
			value := strings.TrimSpace(attr[i+1:])
			if key != "cluster" {
				attrMap[key] = value
			} else {
				// collect cluster info(s)
				// format: "<cluster-id>;<influxdb-info>;<grafana-info>"
				data := strings.Split(value, ";")
				if len(data) >= 3 {
					clusters = append(clusters, ClusterInfo{data[0], data[1], data[2]})
				}
			}
			//scope.Info(fmt.Sprintf("authentication.Authenticate: attributes is %v", authUserInfo.Attributes))
		}
	}

	if value, ok := attrMap["firstName"]; ok {
		authUserInfo.FirstName = value
	}
	if value, ok := attrMap["lastName"]; ok {
		authUserInfo.LastName = value
	}
	if value, ok := attrMap["namespace"]; ok {
		authUserInfo.Namespace = value
	}
	if value, ok := attrMap["company"]; ok {
		authUserInfo.Company = value
	}
	if value, ok := attrMap["url"]; ok {
		authUserInfo.URL = value
	}
	if value, ok := attrMap["agentAccount"]; ok {
		authUserInfo.AgentAccount = value
	}
	if value, ok := attrMap["agentPassword"]; ok {
		authUserInfo.AgentPassword = value
	}
	if value, ok := attrMap["certificate"]; ok {
		authUserInfo.Certificate = value
	}
	if value, ok := attrMap["status"]; ok {
		authUserInfo.Status = value
	}
	if value, ok := attrMap["createdAt"]; ok {
		authUserInfo.CreatedAt = value
	}
	if value, ok := attrMap["updatedAt"]; ok {
		authUserInfo.UpdatedAt = value
	}
	if value, ok := attrMap["phone"]; ok {
		authUserInfo.Phone = value
	}
	if value, ok := attrMap["sendConfirmCount"]; ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			authUserInfo.SendConfirmCount = i
		}
	}
	if value, ok := attrMap["timezone"]; ok {
		authUserInfo.Timezone = value
	}
	if len(clusters) > 0 {
		authUserInfo.Clusters = clusters
	}
}

func (c *AuthLdap) genDescriptionAttrs(authUserInfo *AuthUserInfo) []string {
	attrs := []string{}

	attrs = append(attrs, fmt.Sprintf("firstName=%s", authUserInfo.FirstName))
	attrs = append(attrs, fmt.Sprintf("lastName=%s", authUserInfo.LastName))
	attrs = append(attrs, fmt.Sprintf("namespace=%s", authUserInfo.Namespace))
	attrs = append(attrs, fmt.Sprintf("company=%s", authUserInfo.Company))
	attrs = append(attrs, fmt.Sprintf("url=%s", authUserInfo.URL))
	attrs = append(attrs, fmt.Sprintf("agentAccount=%s", authUserInfo.AgentAccount))
	attrs = append(attrs, fmt.Sprintf("agentPassword=%s", authUserInfo.AgentPassword))
	attrs = append(attrs, fmt.Sprintf("certificate=%s", authUserInfo.Certificate))
	attrs = append(attrs, fmt.Sprintf("status=%s", authUserInfo.Status))
	attrs = append(attrs, fmt.Sprintf("createdAt=%s", authUserInfo.CreatedAt))
	attrs = append(attrs, fmt.Sprintf("updatedAt=%s", authUserInfo.UpdatedAt))
	attrs = append(attrs, fmt.Sprintf("phone=%s", authUserInfo.Phone))
	attrs = append(attrs, fmt.Sprintf("sendConfirmCount=%d", authUserInfo.SendConfirmCount))
	attrs = append(attrs, fmt.Sprintf("timezone=%s", authUserInfo.Timezone))
	if len(authUserInfo.Clusters) > 0 {
		for _, cluster := range authUserInfo.Clusters {
			attrs = append(attrs, fmt.Sprintf("cluster=%s;%s;%s", cluster.ID, cluster.InfluxdbInfo, cluster.GrafanaInfo))
		}
	}

	return attrs
}

func (c *AuthLdap) GetUserCountByDomain(ouName string) int {
	scope.Info("authentication.GetUserCount: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetUserCount: %v", err.Error()))
		return -1
	}

	filter := ""

	if ouName == "" {
		filter = "(&(objectClass=inetOrgPerson)(!uid=_adm))"
	} else {
		filter = fmt.Sprintf("(&(ou=%s)(objectClass=inetOrgPerson)(!uid=_adm))", ouName)
	}

	searchRequest := ldap.NewSearchRequest(
		LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		filter,
		[]string{},
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetUserCount: %v", err.Error()))
		return -1
	} else if sr == nil {
		scope.Error(fmt.Sprintf("ou %s is not found", ouName))
		return -1
	}

	return len(sr.Entries)
}

func (c *AuthLdap) GetUserCountByDomainWithCaller(callerInfo *AuthUserInfo, ouName string) int {
	return c.GetUserCountByDomain(ouName)
}

func (c *AuthLdap) GetAllUserCount() int {
	scope.Info("authentication.GetAllUserCount: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetAllUserCount: %v", err.Error()))
		return -1
	}

	dn := LdapBaseDN

	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		"(&(objectClass=inetOrgPerson)(!uid=_adm))",
		[]string{},
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.GetAllUserCount: %v", err.Error()))
		return -1
	} else if sr == nil {
		msg := ""
		if err != nil {
			msg = string(err.Error())
		} else {
			msg = "unable to get ldap search result"
		}
		scope.Error(fmt.Sprintf("authentication.GetAllUserCount: %s", msg))
		return -1
	}

	return len(sr.Entries)
}

func (c *AuthLdap) GetUserMaximum() int {
	return GetUserMaximum()
}

func (c *AuthLdap) SearchOUByUser(uid string) (string, error) {
	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.searchOU: %v", err.Error()))
		return "", err
	}

	searchRequest := ldap.NewSearchRequest(
		LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		fmt.Sprintf("(uid=%s)", uid),
		LdapUserAttributes,
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.searchOU: %v", err.Error()))
		return "", Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		return "", Errors.NewError(Errors.ReasonUserNotFound, uid)
	} else if len(sr.Entries) == 0 {
		return "", Errors.NewError(Errors.ReasonUserNotFound, uid)
	}

	return sr.Entries[0].GetAttributeValues("ou")[0], err
}

func (c *AuthLdap) IsUserExist(userName string) (bool, error) {
	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.searchOU: %v", err.Error()))
		return false, err
	}

	searchRequest := ldap.NewSearchRequest(
		LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		fmt.Sprintf("(uid=%s)", userName),
		LdapUserAttributes,
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.searchOU: %v", err.Error()))
		return false, Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		return false, Errors.NewError(Errors.ReasonUserNotFound, userName)
	} else if len(sr.Entries) == 0 {
		return false, nil
	}

	return true, nil
}

func (c *AuthLdap) GenNewOUName(ouName string) (string, error) {
	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.searchOU: %v", err.Error()))
		return "", err
	}

	//step.1 get all ou
	searchRequest := ldap.NewSearchRequest(
		LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		"(objectClass=organizationalUnit)",
		[]string{"ou"},
		nil)

	sr, err := lconn.Search(searchRequest)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.genNewOU: %v", err.Error()))
		return "", Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		scope.Error("authentication.genNewOU failed")
		return "", Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	}

	//step.2 find match ou
	ouNumList := []int{}
	for _, entry := range sr.Entries {
		tempOU := entry.GetAttributeValues("ou")[0]
		index := strings.Index(tempOU, ouName+"-")

		if index == 0 {
			tempSplit := strings.Split(tempOU, ouName+"-")[1]
			tempNum, err := strconv.Atoi(tempSplit)
			if err == nil {
				ouNumList = append(ouNumList, tempNum)
			}
		}
	}

	sort.Ints(ouNumList)

	//step.3 add number
	retOu := ouName
	if len(ouNumList) > 0 {
		retOu = ouName + fmt.Sprintf("-%d", ouNumList[len(ouNumList)-1]+1)
	} else {

	}

	return retOu, nil
}

func (c *AuthLdap) GetUserListByStatus(authUserInfoList *[]AuthUserInfo, ou string, status string) error {
	scope.Info("authentication.GetUserListByStatus: starting...")

	lconn, err := c.adminLoginLDAP()
	if lconn != nil {
		defer lconn.Close()
	}
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadUserList: %v", err.Error()))
		return err
	}

	tempBaseDN := ""
	tempFilter := ""

	if ou == "" {
		tempBaseDN = LdapBaseDN
		tempFilter = "(&(objectClass=inetOrgPerson))"
	} else {
		tempBaseDN = fmt.Sprintf("ou=%s,%s", ou, LdapBaseDN)
		tempFilter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(description=status=%s))", status)
	}

	searchRequest := ldap.NewSearchRequest(
		tempBaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		tempFilter,
		LdapUserAttributes,
		nil)

	sr, err := lconn.Search(searchRequest)

	if err != nil {
		scope.Error(fmt.Sprintf("authentication.ReadUserList: %v", err.Error()))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	} else if sr == nil {
		scope.Error(fmt.Sprintf("ou %s is not found", ou))
		return Errors.NewError(Errors.ReasonFailedToReadDB, "ldap")
	}

	for _, entry := range sr.Entries {
		tempUserInfo := AuthUserInfo{}
		c.loadUserAllData(&tempUserInfo, entry)
		*authUserInfoList = append(*authUserInfoList, tempUserInfo)
	}

	scope.Info(fmt.Sprintf("authentication.GetUserListByStatus: read user by Ou successfully."))

	return nil
}
