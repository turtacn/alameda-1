package accounts

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServiceAccount struct {
	Config *APIServerConfig.Config
}

func NewServiceAccount(cfg *APIServerConfig.Config) *ServiceAccount {
	service := ServiceAccount{}
	service.Config = cfg
	return &service
}

func (c *ServiceAccount) Authorize(user, loginUser, loginDomain, loginRole string) error {
	if loginUser == "" || loginDomain == "" {
		return Errors.NewError(Errors.ReasonInvalidParams)
	}
	if loginUser == "super" && loginRole == "super" {
		return nil
	}
	if user != "" {
		if user == loginUser {
			return nil
		} else {
			if loginRole == RoleDomainAdmin || loginRole == RoleSuper {
				return nil
			} else {
				return Errors.NewError(Errors.ReasonNotAuthorized)
			}
		}
	} else {
		return Errors.NewError(Errors.ReasonInvalidParams)
	}
}

func (c *ServiceAccount) CreateUser(ctx context.Context, in *Accounts.CreateUserRequest) (*Accounts.CreateUserResponse, error) {
	scope.Debug("Request received from CreateUser grpc function: " + AlamedaUtils.InterfaceToString(in))

	authInfo, err := Authenticate(ctx)
	if err != nil {
		return &Accounts.CreateUserResponse{}, err
	}
	err = c.Authorize(in.Name, authInfo.Name, authInfo.DomainName, authInfo.Role)
	if err != nil {
		return &Accounts.CreateUserResponse{}, err
	}
	userSvc := ServiceUser{Config: c.Config}
	out, err := userSvc.CreateUser(authInfo, in)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *ServiceAccount) ReadUser(ctx context.Context, in *Accounts.ReadUserRequest) (*Accounts.ReadUserResponse, error) {
	scope.Debug("Request received from ReadUser grpc function: " + AlamedaUtils.InterfaceToString(in))

	authInfo, err := Authenticate(ctx)
	if err != nil {
		return &Accounts.ReadUserResponse{}, err
	}
	err = c.Authorize(in.Name, authInfo.Name, authInfo.DomainName, authInfo.Role)
	if err != nil {
		return &Accounts.ReadUserResponse{}, err
	}
	userSvc := ServiceUser{Config: c.Config}
	out, err := userSvc.ReadUser(authInfo, in)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *ServiceAccount) UpdateUser(ctx context.Context, in *Accounts.UpdateUserRequest) (*Accounts.UpdateUserResponse, error) {
	scope.Debug("Request received from UpdateUser grpc function: " + AlamedaUtils.InterfaceToString(in))

	authInfo, err := Authenticate(ctx)
	if err != nil {
		return &Accounts.UpdateUserResponse{}, err
	}
	err = c.Authorize(in.Name, authInfo.Name, authInfo.DomainName, authInfo.Role)
	if err != nil {
		return &Accounts.UpdateUserResponse{}, err
	}
	userSvc := ServiceUser{Config: c.Config}
	out, err := userSvc.UpdateUser(authInfo, in)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *ServiceAccount) DeleteUser(ctx context.Context, in *Accounts.DeleteUserRequest) (*Accounts.DeleteUserResponse, error) {
	scope.Debug("Request received from DeleteUser grpc function: " + AlamedaUtils.InterfaceToString(in))

	authInfo, err := Authenticate(ctx)
	if err != nil {
		return &Accounts.DeleteUserResponse{}, err
	}
	err = c.Authorize(in.Name, authInfo.Name, authInfo.DomainName, authInfo.Role)
	if err != nil {
		return &Accounts.DeleteUserResponse{}, err
	}
	if in.Name == "super" {
		return &Accounts.DeleteUserResponse{}, Errors.NewError(Errors.ReasonInvalidRequest, "Cannot delete user(super)")
	}
	userSvc := ServiceUser{Config: c.Config}
	out, err := userSvc.DeleteUser(authInfo, in)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *ServiceAccount) Login(ctx context.Context, in *Accounts.LoginRequest) (*Accounts.LoginResponse, error) {
	scope.Debug("Request received from Login grpc function: " + AlamedaUtils.InterfaceToString(in))

	userSvc := ServiceUser{Config: c.Config}
	out, err := userSvc.UserLogin(in)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (c *ServiceAccount) Logout(ctx context.Context, in *Accounts.LogoutRequest) (*status.Status, error) {
	scope.Debug("Request received from Logout grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
