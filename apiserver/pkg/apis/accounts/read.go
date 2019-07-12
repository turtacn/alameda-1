package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	// Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
)

func (c *ServiceUser) ReadUser(caller *entity.User, in *Accounts.ReadUserRequest) (*Accounts.ReadUserResponse, error) {
	scope.Debug("[apis.accounts.ReadUser]")

	response := Accounts.ReadUserResponse{}

	owner := authentication.NewAuthUserInfo("", in.Name)
	err := caller.ReadUser(owner)
	if err == nil {
		response.Name = owner.Name
		response.DomainName = owner.DomainName
		response.Role = owner.Role
		response.Company = owner.Company
		response.Email = owner.Email
		response.FirstName = owner.FirstName
		response.LastName = owner.LastName
		response.Phone = owner.Phone
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
		}
		return &response, nil
	} else {
		scope.Errorf("Failed to read user(%s) info: %s", in.Name, err.Error())
		// return &response, Errors.NewError(Errors.ReasonUserReadFailed, in.Name)
		return &response, err
	}
}
