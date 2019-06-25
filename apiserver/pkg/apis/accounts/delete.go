package accounts

import (
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
)

func (c *ServiceUser) DeleteUser(ctx context.Context, in *Accounts.DeleteUserRequest) (*Accounts.DeleteUserResponse, error) {
	scope.Debug("[apis.accounts.DeleteUser]")

	response := Accounts.DeleteUserResponse{}
	dinfo := authentication.NewAuthUserInfo(in.Name, in.DomainName)
	token := in.Token

	if token != "" {
		_, err := authentication.Validate(dinfo, token)
		if err != nil {
			scope.Error(err.Error())
			return &response, err
		}
		caller := entity.User{}
		caller.Info.Name = dinfo.Name
		caller.Info.DomainName = dinfo.DomainName
		caller.Info.Token = token
		caller.Info.Cookie = dinfo.Cookie
		caller.Info.Role = dinfo.Role
		caller.Config = c.Config
		if dinfo.Role == RoleSuper || (dinfo.Role == RoleDomainAdmin && dinfo.DomainName == in.DomainName) {
			isExist, err := caller.IsUserExist(in.Name)
			if err != nil {
				scope.Errorf("Failed to delete user(%s), unable to check user: %s", in.Name, err.Error())
				return &response, err
			}
			// TODO: delete (or reserve) influxdb and grafana container for the user
			owner := authentication.NewAuthUserInfo(in.DomainName, in.Name)
			if isExist {
				err = caller.DeleteUser(owner)
				if err != nil {
					scope.Errorf("Failed to delete user(%s) from domain(%s): %s", owner.Name, owner.DomainName, err.Error())
					return &response, err
				}
				scope.Infof("Delete user(%s) from domain(%s) successfully", owner.Name, owner.DomainName)
			} else {
				scope.Warnf("user(%s) does not exist!", in.Name)
			}
			response.Name = owner.Name
			response.DomainName = owner.DomainName
			response.Role = owner.Role
			return &response, nil
		} else {
			return &response, Errors.NewError(Errors.ReasonInvalidRequest, fmt.Sprintf("no permission to delete user(%s)", in.Name))
		}
	} else {
		return &response, Errors.NewError(Errors.ReasonInvalidCredential)
	}
}
