package accounts

import (
	"context"
	"errors"
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	"github.com/containers-ai/alameda/apiserver/pkg/config"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

const (
	RoleSuper       = "super"
	RoleDomainAdmin = "domain-admin"
	RoleUser        = "user"
)

type ServiceUser struct {
	Arch   string
	Config *config.Config
}

func NewServiceUser(arch string) *ServiceUser {
	scope.Debug("NewServiceUser")

	return &ServiceUser{Arch: arch}
}

func (c *ServiceUser) IsUserExist(userName string) (bool, error) {
	user := entity.NewUserEntity(userName, "", "", c.Config)

	exist, err := user.IsUserExist(userName)
	if err != nil {
		return exist, err
	}

	return exist, nil
}

func Authenticate(ctx context.Context) (authentication.AuthUserInfo, error) {
	authInfo := authentication.AuthUserInfo{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		scope.Errorf("No authentication info contained in context")
		err := Errors.NewError(Errors.ReasonInvalidParams)
		return authInfo, err
	}
	username := strings.Join(md.Get("Username"), "")
	password := strings.Join(md.Get("Password"), "")
	token := strings.Join(md.Get("Token"), "")
	if token != "" {
		// authenticate user by token
		_, err := authentication.Validate(&authInfo, token)
		if err != nil {
			return authInfo, err
		}
	} else {
		if username != "" && password != "" {
			// authenticate user by username/password
			userInfo := entity.NewUserEntity(username, "", "", nil)
			err := userInfo.Authenticate(password)
			if err != nil {
				scope.Errorf("Failed to authenticate user(%s): %s", username, err.Error())
				return authInfo, err
			}
			authInfo.Name = userInfo.Info.Name
			authInfo.DomainName = userInfo.Info.DomainName
			authInfo.Role = userInfo.Info.Role
			authInfo.Token = userInfo.Info.Token
			authInfo.Cookie = userInfo.Info.Cookie
		} else {
			err := errors.New("invalid authentication data provided")
			return authInfo, err
		}
	}
	return authInfo, nil
}

func GetUserCredentialFromContext(ctx context.Context) (string, string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		scope.Errorf("No authentication info contained in context")
		err := Errors.NewError(Errors.ReasonInvalidParams)
		return "", "", "", err
	}
	username := strings.Join(md.Get("Username"), "")
	password := strings.Join(md.Get("Password"), "")
	token := strings.Join(md.Get("Token"), "")
	return username, password, token, nil
}

func CreateFakeUserContainers(ainfo *authentication.AuthUserInfo) (string, string, error) {
	tm := time.Now().UnixNano()
	influxdbInfo := fmt.Sprintf("influxdb.%d.federatorai:8086", tm)
	grafanaInfo := fmt.Sprintf("grafana.%d.federatorai:3000", tm)
	scope.Infof("==>XXX create container for user(%s)", ainfo.Name)
	return influxdbInfo, grafanaInfo, nil
}

func DeleteFakeUserContainers(ainfo authentication.AuthUserInfo) error {
	user := ainfo.Name
	influxdbInfo := ainfo.InfluxdbInfo
	grafanaInfo := ainfo.GrafanaInfo
	scope.Infof("==>XXX delete user(%s) container:", user)
	scope.Infof("\tInfluxDB: %s", influxdbInfo)
	scope.Infof("\tGrafana : %s", grafanaInfo)
	return nil
}
