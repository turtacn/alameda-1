package accounts

import (
	"fmt"
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
)

func (c *ServiceUser) ReadUser(ctx context.Context, in *Accounts.ReadUserRequest) (*Accounts.ReadUserResponse, error) {
	scope.Debug("[apis.accounts.ReadUser]")

	response := Accounts.ReadUserResponse{}
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
		caller.Info.Token = ainfo.Token
		caller.Info.Cookie = ainfo.Cookie
		caller.Info.Role = ainfo.Role
		caller.Config = c.Config
		if (ainfo.Name != in.Name && (ainfo.Role == RoleDomainAdmin || ainfo.Role == RoleSuper)) || ainfo.Name == in.Name {
			owner := authentication.NewAuthUserInfo(in.DomainName, in.Name)
			err = caller.ReadUser(owner)
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
				response.InfluxdbInfo = owner.InfluxdbInfo
				response.GrafanaInfo = owner.GrafanaInfo
				return &response, nil
			} else {
				//scope.Errorf("Failed to read user(%s) info: %s", in.Name, err.Error())
				return &response, Errors.NewError(Errors.ReasonUserReadFailed, in.Name)
			}
		} else {
			return &response, Errors.NewError(Errors.ReasonInvalidRequest, fmt.Sprintf("no permission to read user(%s) info", in.Name))
		}
	} else {
		return &response, Errors.NewError(Errors.ReasonInvalidCredential)
	}
}
