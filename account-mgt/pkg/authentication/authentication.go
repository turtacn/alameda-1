package authentication

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	//Logger "github.com/prophetstor-ai/common-lib/pkg/log"
	"crypto/tls"
	Config "github.com/containers-ai/alameda/pkg/utils/conf"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"net/http"
)

var AuthType = "ldap"
var TokenType = "JWT"
var TokenDuration = 3600
var PlatformType = "platform"

var UserMaximum = 0

var scope = log.RegisterScope("account-mgt", "account-mgt log", 0)

type AuthUserInfo struct {
	Name       string
	DomainName string
	Password   string
	Email      string
	Role       string
	ID         string

	FirstName     string
	LastName      string
	Namespace     string
	Company       string
	URL           string
	AgentAccount  string
	AgentPassword string
	Certificate   string
	Phone         string
	Status        string
	CreatedAt     string
	UpdatedAt     string

	SendConfirmCount int
	Timezone         string

	DN          string
	Cookie      string
	CookieValue string
	Token       string
}

func NewAuthUserInfo(domainName, name string) *AuthUserInfo {
	tempDn := ""
	if domainName == "" {
		tempDn = fmt.Sprintf("uid=%s,%s", name, LdapBaseDN)
	} else {
		tempDn = fmt.Sprintf("uid=%s,ou=%s,%s", name, domainName, LdapBaseDN)
	}

	newUserInfo := &AuthUserInfo{
		Name:       name,
		Password:   "",
		DomainName: domainName,
		Role:       "",
		Email:      "",
		ID:         "",

		FirstName:     "",
		LastName:      "",
		Namespace:     "",
		Company:       "",
		URL:           "",
		AgentAccount:  "",
		AgentPassword: "",
		Certificate:   "",
		Phone:         "",
		Status:        "",
		CreatedAt:     "",
		UpdatedAt:     "",

		SendConfirmCount: 0,
		Timezone:         "",

		DN:     tempDn,
		Cookie: "",
		Token:  "",
	}

	return newUserInfo
}

type AuthInterface interface {
	Authenticate(authUserInfo *AuthUserInfo) (string, error)
	Validate(authUserInfo *AuthUserInfo, token string) (string, error)
	ChangePassword(authUserInfo *AuthUserInfo, newPassword string) error
	CreateUser(authUserInfo *AuthUserInfo, joinDomain bool) error
	ReadUser(authUserInfo *AuthUserInfo) error
	UpdateUser(authUserInfo *AuthUserInfo) error
	DeleteUser(authUserInfo *AuthUserInfo) error

	CreateUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo, joinDomain bool) error
	ReadUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error
	UpdateUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error
	DeleteUserWithCaller(callerInfo *AuthUserInfo, ownerInfo *AuthUserInfo) error

	GetUserListByDomain(ownerInfoList *[]AuthUserInfo, domain string, limit int, page int) error
	GetUserListByDomainWithCaller(callerInfo *AuthUserInfo, ownerInfoList *[]AuthUserInfo, domain string, limit int, page int) error
	GetUserCountByDomain(ouName string) int
	GetUserCountByDomainWithCaller(callerInfo *AuthUserInfo, ouName string) int
	GetAllUserCount() int

	GetUserMaximum() int
	GetUserListByStatus(authUserInfoList *[]AuthUserInfo, domain string, status string) error

	IsUserExist(userName string) (bool, error)
}

func NewAuthenticationClient(aNamespace string) AuthInterface {
	switch AuthType {
	case "ldap":
		client := AuthLdap{}
		return &client
	default:
		break
	}
	return nil
}

type AuthenticationConfig struct {
	Addr            string
	Port            uint16
	AuthCategory    string
	TokenCategory   string
	PlatformType    string
	TokenExpiration int
}

func AuthenticationInit(addr string, port uint16, authCategory, tokenCategory, platformType string, tokenExpiration int) error {
	LdapIP = addr
	LdapPort = port

	PlatformType = platformType

	AuthType = authCategory
	TokenType = tokenCategory
	TokenDuration = tokenExpiration

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return nil
}

func AuthenticationInitWithConfig(config AuthenticationConfig) error {
	LdapIP = config.Addr
	LdapPort = config.Port

	PlatformType = config.PlatformType

	AuthType = config.AuthCategory
	TokenType = config.TokenCategory
	TokenDuration = config.TokenExpiration

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return nil
}

func AuthenticationInitWithConfigFile(path string) {
	//path = "/etc/account-mgt/account-mgt.toml"
	Config.ConfigInit(path)

	category := Config.Get("general.arch", "platform").(string)
	addr := Config.Get("authentication.address", "ldap.fed-account").(string)

	port := uint16(Config.Get("authentication.port", int64(389)).(int64))
	authType := Config.Get("authentication.auth_type", "ldap").(string)
	tokenType := Config.Get("authentication.token_type", "JWT").(string)
	tokenExpiration := int(Config.Get("authentication.token_expiration", int64(3600)).(int64))
	userMaximum := int(Config.Get("authentication.user_maximum", int64(1000)).(int64))
	//logPath := Config.Get("logger.file", "/var/log/account-mgt.log").(string)
	//Logger.LoggerInit(logPath)
	opt := log.DefaultOptions()
	opt.RotationMaxSize = 100
	opt.RotationMaxBackups = 7
	opt.RotateOutputPath = Config.Get("logger.file", "/var/log/account-mgt.log").(string)
	err := log.Configure(opt)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %s", err.Error())
	} else {
		// re-register scope due to apply the new config parameters
		scope = log.RegisterScope("accmgt-authentication", "account-mgt authentication", 0)
	}

	LdapIP = addr
	LdapPort = port

	PlatformType = category

	AuthType = authType
	TokenType = tokenType
	TokenDuration = tokenExpiration
	UserMaximum = userMaximum

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	scope.Info(fmt.Sprintf("authentication.Init: category=%s", category))
	scope.Info(fmt.Sprintf("authentication.Init: authType=%s", authType))
}

func GetAuthType() string {
	return AuthType
}

func GetTokenType() string {
	return TokenType
}

func GetTokenDuration() int {
	return TokenDuration
}

func GetUserMaximum() int {
	return UserMaximum
}

func passwordSshaEncode(password string) (string, error) {
	salt := make([]byte, 4)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	pass := []byte(password)
	str := append(pass[:], salt[:]...)
	sum := sha1.Sum(str)
	result := append(sum[:], salt[:]...)
	ret := fmt.Sprintf("{SSHA}%s", base64.StdEncoding.EncodeToString(result))
	return ret, nil
}

func Validate(authUserInfo *AuthUserInfo, token string) (string, error) {
	scope.Info("authentication.Validate: starting...")
	// First step is to confirm user has not logged out
	claims, newToken, err := JwtDecode(token)
	if err != nil {
		scope.Error(fmt.Sprintf("authentication.Validate: %v", err.Error()))
		return token, err
	}
	for k, _ := range claims {
		switch k {
		case "name":
			authUserInfo.Name = claims["name"]
		case "domainName":
			authUserInfo.DomainName = claims["domainName"]
		case "namespace":
			authUserInfo.Namespace = claims["namespace"]
		case "role":
			authUserInfo.Role = claims["role"]
		case "cookie":
			authUserInfo.Cookie = claims["cookie"]
		default:
		}
	}

	authUserInfo.DN = fmt.Sprintf("uid=%s,ou=%s,%s", authUserInfo.Name, authUserInfo.DomainName, LdapBaseDN)
	scope.Info(fmt.Sprintf("authentication.Validate: orgJWT=%s", token))
	scope.Info(fmt.Sprintf("authentication.Validate: newJWT=%s", newToken))

	return newToken, nil
}
