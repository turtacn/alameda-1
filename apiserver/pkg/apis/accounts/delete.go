package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
)

func (c *ServiceUser) DeleteUser(caller *entity.User, in *Accounts.DeleteUserRequest) (*Accounts.DeleteUserResponse, error) {
	scope.Debug("[apis.accounts.DeleteUser]")

	response := Accounts.DeleteUserResponse{}

	isExist, err := caller.IsUserExist(in.Name)
	if err != nil {
		scope.Errorf("Failed to delete user(%s), unable to check user: %s", in.Name, err.Error())
		return &response, err
	}

	owner := authentication.NewAuthUserInfo("", in.Name)
	if isExist {
		err := caller.ReadUser(owner)
		if err != nil {
			scope.Errorf("Failed to read user(%s) info from ldap for user deletion: %s", owner.Name, err.Error())
			return &response, err
		}
		// TODO: delete (or reserve) influxdb and grafana container for the user
		err = DeleteFakeUserContainers(*owner)
		if err != nil {
			scope.Errorf("Failed to remove user(%s) containers: %s", owner.Name, err.Error())
			return &response, err
		}
		err = caller.DeleteUser(owner)
		if err != nil {
			scope.Errorf("Failed to delete user(%s) from domain(%s): %s", owner.Name, owner.DomainName, err.Error())
			return &response, err
		}
		scope.Infof("Delete user(%s) from domain(%s) successfully", owner.Name, owner.DomainName)
	} else {
		scope.Warnf("user(%s) does not exist!", owner.Name)
	}
	response.Name = owner.Name
	response.DomainName = owner.DomainName
	response.Role = owner.Role
	return &response, nil
}
