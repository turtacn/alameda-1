package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
)

func (c *ServiceUser) CreateUser(caller *entity.User, in *Accounts.CreateUserRequest) (*Accounts.CreateUserResponse, error) {
	scope.Debug("[apis.accounts.CreateUser]")

	response := Accounts.CreateUserResponse{}
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
	if len(in.Clusters) > 0 {
		for _, cluster := range in.Clusters {
			cinfo := authentication.ClusterInfo{ID: cluster.ID, InfluxdbInfo: cluster.InfluxdbInfo, GrafanaInfo: cluster.GrafanaInfo}
			owner.Clusters = append(owner.Clusters, cinfo)
		}
	} else {
		owner.Clusters = []authentication.ClusterInfo{}
	}
	err := caller.CreateUser(owner)
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
		if len(owner.Clusters) > 0 {
			for _, cluster := range owner.Clusters {
				cinfo := new(Accounts.ClusterInfo)
				cinfo.ID = cluster.ID
				cinfo.InfluxdbInfo = cluster.InfluxdbInfo
				cinfo.GrafanaInfo = cluster.GrafanaInfo
				response.Clusters = append(response.Clusters, cinfo)
			}
		} else {
			response.Clusters = []*Accounts.ClusterInfo{}
		}
		return &response, nil
	} else {
		scope.Errorf("Failed to create user(%s) in domain(%s): %s", owner.Name, owner.DomainName, err.Error())
		return &response, err
	}
}
