package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
)

func (c *ServiceUser) CreateUser(authInfo authentication.AuthUserInfo, in *Accounts.CreateUserRequest) (*Accounts.CreateUserResponse, error) {
	scope.Debug("[apis.accounts.CreateUser]")

	response := Accounts.CreateUserResponse{}
	caller := entity.User{}
	caller.Info.Name = authInfo.Name
	caller.Info.DomainName = authInfo.DomainName
	caller.Info.Token = authInfo.Token
	caller.Info.Cookie = authInfo.Cookie
	caller.Info.Role = authInfo.Role
	caller.Config = c.Config

	// owner := authentication.NewAuthUserInfo(in.DomainName, in.Name)
	// only one domain "localdomain"
	owner := authentication.NewAuthUserInfo("localdomain", in.Name)
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
		err1 := DeleteFakeUserContainers(*owner)
		if err1 != nil {
			scope.Errorf("Failed to rollback containers for user(%s) creation failure: %s", owner.Name, err1.Error())
		}
		return &response, err
	}
}
