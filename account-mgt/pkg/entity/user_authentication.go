package entity

import (
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var scope = log.RegisterScope("accmgt-entity", "account-mgt entity", 0)

func (c *User) Authenticate(password string) error {
	domainName := c.Info.DomainName
	name := c.Info.Name

	authUserInfo := authentication.NewAuthUserInfo(domainName, name)
	authUserInfo.Password = password
	authUserInfo.Cookie = c.Info.Cookie

	c.Info.Password = password

	client := authentication.NewAuthenticationClient(c.Config)
	token, err := client.Authenticate(authUserInfo)
	if err != nil {
		scope.Error(fmt.Sprintf("User.Authenticate: %v", err.Error()))
		return err
	}

	c.Info.Password = password
	c.Info.Token = token
	c.Info.Namespace = authUserInfo.Namespace
	c.Info.DomainName = authUserInfo.DomainName
	c.Info.Role = authUserInfo.Role

	c.Info.Password = authUserInfo.Password
	c.Info.Cookie = authUserInfo.Cookie
	c.Info.CookieValue = authUserInfo.CookieValue

	return nil
}

func (c *User) DoAuthentication() error {
	if c.Info.Token != "" {
		return c.Validate()
	} else if c.Info.Name != "" && c.Info.Password != "" {
		return c.Authenticate(c.Info.Password)
	} else {
		return Errors.NewError(Errors.ReasonInvalidCredential)
	}
}

func (c *User) Validate() error {
	domainName := c.Info.DomainName
	name := c.Info.Name

	authUserInfo := authentication.NewAuthUserInfo(domainName, name)
	authUserInfo.Cookie = c.Info.Cookie

	client := authentication.NewAuthenticationClient(c.Config)
	newToken, err := client.Validate(authUserInfo, c.Info.Token)
	if err != nil {
		scope.Error(fmt.Sprintf("User.Validate: %v", err.Error()))
		return err
	}
	c.Info.Name = authUserInfo.Name
	c.Info.Namespace = authUserInfo.Namespace
	c.Info.DomainName = authUserInfo.DomainName
	c.Info.Role = authUserInfo.Role
	c.Info.Token = newToken

	c.Info.Password = authUserInfo.Password
	c.Info.Cookie = authUserInfo.Cookie
	c.Info.CookieValue = authUserInfo.CookieValue

	return nil
}

func (c *User) ChangePassword(authUserInfo *authentication.AuthUserInfo, newPassword string) error {
	client := authentication.NewAuthenticationClient(c.Config)
	err := client.ChangePassword(authUserInfo, newPassword)
	if err != nil {
		scope.Error(fmt.Sprintf("User.UpdateUser: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) CreateUser(ownerInfo *authentication.AuthUserInfo) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.CreateUserWithCaller(callInfo, ownerInfo, true)
	if err != nil {
		scope.Error(fmt.Sprintf("User.CreateUser: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) ReadUser(ownerInfo *authentication.AuthUserInfo) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.ReadUserWithCaller(callInfo, ownerInfo)
	if err != nil {
		scope.Error(fmt.Sprintf("User.ReadUser: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) ReadUserList(authUserInfoList *[]authentication.AuthUserInfo, limit int, page int) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	domain := ""
	if callInfo.Role != "super" {
		domain = callInfo.DomainName
	}

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.GetUserListByDomainWithCaller(callInfo, authUserInfoList, domain, limit, page)
	if err != nil {
		scope.Error(fmt.Sprintf("User.ReadUserListByDomain: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) ReadUserListByDomain(authUserInfoList *[]authentication.AuthUserInfo, domain string, limit int, page int) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.GetUserListByDomainWithCaller(callInfo, authUserInfoList, domain, limit, page)
	if err != nil {
		scope.Error(fmt.Sprintf("User.ReadUserListByDomain: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) UpdateUser(ownerInfo *authentication.AuthUserInfo) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.UpdateUserWithCaller(callInfo, ownerInfo)
	if err != nil {
		scope.Error(fmt.Sprintf("User.UpdateUser: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) DeleteUser(ownerInfo *authentication.AuthUserInfo) error {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)
	err := client.DeleteUserWithCaller(callInfo, ownerInfo)
	if err != nil {
		scope.Error(fmt.Sprintf("User.DeleteUser: %v", err.Error()))
		return err
	}

	return nil
}

func (c *User) IsUserExist(userName string) (bool, error) {
	client := authentication.NewAuthenticationClient(c.Config)
	exist, err := client.IsUserExist(userName)
	if err != nil {
		scope.Error(fmt.Sprintf("User.IsUserExist: %v", err.Error()))
		return exist, err
	}

	return exist, nil
}

func (c *User) GetUserCount() int {
	callInfo := authentication.NewAuthUserInfo(c.Info.DomainName, c.Info.Name)
	callInfo.Password = c.Info.Password
	callInfo.Cookie = c.Info.Cookie
	callInfo.Role = c.Info.Role

	client := authentication.NewAuthenticationClient(c.Config)

	//tempDomain := ""
	//if c.Info.Role != "super" {
	//	tempDomain = c.Info.DomainName
	//}

	return client.GetUserCountByDomainWithCaller(callInfo, c.Info.DomainName)
}

func (c *User) GetAllUserCount() int {
	client := authentication.NewAuthenticationClient(c.Config)
	return client.GetAllUserCount()
}
