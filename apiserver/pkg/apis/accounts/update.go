package accounts

import (
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
)

func (c *ServiceUser) UpdateUser(ctx context.Context, in *Accounts.UpdateUserRequest) (*Accounts.UpdateUserResponse, error) {
	scope.Debug("[apis.accounts.UpdateUser]")

	response := Accounts.UpdateUserResponse{}
	uinfo := authentication.AuthUserInfo{}
	token := in.Token

	if token != "" {
		_, err := authentication.Validate(&uinfo, token)
		if err != nil {
			scope.Error(err.Error())
			return &response, err
		}
		caller := entity.User{}
		caller.Info.Name = uinfo.Name
		caller.Info.DomainName = uinfo.DomainName
		caller.Info.Token = token
		caller.Info.Cookie = uinfo.Cookie
		caller.Info.Role = uinfo.Role
		caller.Config = c.Config
		owner := authentication.NewAuthUserInfo("", in.Name)
		err = caller.ReadUser(owner)
		if err != nil {
			scope.Errorf("Failed to read user(%s) info for update: %s", owner.Name, err.Error())
			return &response, err
		}
		if uinfo.Role == RoleSuper || (uinfo.Role == RoleDomainAdmin && uinfo.DomainName == owner.DomainName) || uinfo.Name == owner.Name {
			if in.FirstName != "" {
				owner.FirstName = in.FirstName
			}
			if in.LastName != "" {
				owner.LastName = in.LastName
			}
			if in.Role != "" {
				// change user role only if caller role is super or domain-admin
				if in.Role == RoleDomainAdmin && (uinfo.Role == RoleSuper || uinfo.Role == RoleDomainAdmin) {
					owner.Role = in.Role
				}
			}
			if in.Company != "" {
				owner.Company = in.Company
			}
			if in.Email != "" {
				owner.Email = in.Email
			}
			if in.Password != "" {
				owner.Password = in.Password
			}
			if in.URL != "" {
				owner.URL = in.URL
			}
			if in.Phone != "" {
				owner.Phone = in.Phone
			}
			err = caller.UpdateUser(owner)
			if err != nil {
				scope.Errorf("Failed to update user(%s) account info: %s", owner.Name, err.Error())
				return &response, err
			} else {
				scope.Infof("Update user(%s) account info successfully", owner.Name)
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
				response.InfluxdbInfo = owner.InfluxdbInfo
				response.GrafanaInfo = owner.GrafanaInfo
				return &response, nil
			}
		} else {
			return &response, Errors.NewError(Errors.ReasonInvalidRequest, fmt.Sprintf("no permission to update user(%s) info", in.Name))
		}
	} else {
		return &response, Errors.NewError(Errors.ReasonInvalidCredential)
	}
}
