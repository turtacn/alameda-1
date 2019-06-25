package accounts

import (
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
	"time"
)

func CreateFakeUserContainers(ainfo *authentication.AuthUserInfo) (string, string, error) {
	tm := time.Now().UnixNano()
	influxdbInfo := fmt.Sprintf("influxdb.%d.federatorai:8086", tm)
	grafanaInfo := fmt.Sprintf("grafana.%d.federatorai:3000", tm)
	return influxdbInfo, grafanaInfo, nil
}

func (c *ServiceUser) CreateUser(ctx context.Context, in *Accounts.CreateUserRequest) (*Accounts.CreateUserResponse, error) {
	scope.Debug("[apis.accounts.CreateUser]")

	response := Accounts.CreateUserResponse{}
	ainfo := authentication.AuthUserInfo{}
	token := in.Token

	if token != "" {
		_, err := authentication.Validate(&ainfo, token)
		if err != nil {
			scope.Error(err.Error())
			return &response, err
		}
		caller := entity.User{}
		caller.Info.Name = ainfo.Name
		caller.Info.DomainName = ainfo.DomainName
		caller.Info.Token = token
		caller.Info.Cookie = ainfo.Cookie
		caller.Info.Role = ainfo.Role
		caller.Config = c.Config
		if ainfo.Role == RoleSuper || (ainfo.Role == RoleDomainAdmin && ainfo.DomainName == in.DomainName) {
			owner := authentication.NewAuthUserInfo(in.DomainName, in.Name)
			owner.Password = in.Password
			owner.Role = in.Role
			owner.Company = in.Company
			owner.FirstName = in.FirstName
			owner.LastName = in.LastName
			owner.Phone = in.Phone
			owner.URL = in.URL
			owner.Status = "created"
			// TODO: call to creating influxdb and grafana container functions, continue to create ldap user if success
			influxdbInfo, grafanaInfo, err := CreateFakeUserContainers(owner)
			if err != nil {
				scope.Errorf("Failed to create service container during create user(%s): %s", owner.Name, err.Error())
				return &response, err
			}
			scope.Infof("create user(%s) influxdb: %s", owner.Name, influxdbInfo)
			scope.Infof("create user(%s) grafana: %s", owner.Name, grafanaInfo)
			owner.InfluxdbInfo = influxdbInfo
			owner.GrafanaInfo = grafanaInfo
			err = caller.CreateUser(owner)
			if err == nil {
				scope.Infof("Create user(%s) in domain(%s) successfully", owner.Name, owner.DomainName)
				response.Name = owner.Name
				response.DomainName = owner.DomainName
				response.Role = owner.Role
				response.Company = owner.Company
				response.FirstName = owner.FirstName
				response.LastName = owner.LastName
				response.Phone = owner.Phone
				response.Email = owner.Email
				response.URL = owner.URL
				response.Status = owner.Status
				response.InfluxdbInfo = influxdbInfo
				response.GrafanaInfo = grafanaInfo
				return &response, nil
			} else {
				// TODO: remove the containers which created for this user
				scope.Errorf("Failed to create user(%s) in domain(%s): %s", owner.Name, owner.DomainName, err.Error())
				return &response, err
			}
		} else {
			return &response, Errors.NewError(Errors.ReasonInvalidRequest, fmt.Sprintf("no permission to create user(%s)", in.Name))
		}
	} else {
		return &response, Errors.NewError(Errors.ReasonInvalidCredential)
	}
}
