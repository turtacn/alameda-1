package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/authentication"
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
)

func (c *ServiceUser) UserLogin(ctx context.Context, in *Accounts.LoginRequest) (*Accounts.LoginResponse, error) {
	scope.Debug("[apis.accounts.UserLogin]")

	response := Accounts.LoginResponse{}

	caller := entity.User{}
	name := in.Name
	password := in.Password
	domainName := in.DomainName

	caller.Info.Name = name
	caller.Info.Password = password
	caller.Info.DomainName = domainName
	caller.Config = c.Config

	err := caller.Authenticate(password)
	if err != nil {
		scope.Errorf("Authenticate user (%s) fail: %s", name, err.Error())
		return &response, err
	}

	response.AccessToken = caller.Info.Token
	response.TokenType = authentication.TokenType
	response.ExpireIn = int32(authentication.TokenDuration)

	return &response, nil
}
