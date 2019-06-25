package accounts

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	"github.com/containers-ai/alameda/apiserver/pkg/config"
)

const (
	RoleSuper       = "super"
	RoleDomainAdmin = "domain-admin"
	RoleUser        = "user"
)

type ServiceUser struct {
	Arch   string
	Config *config.Config
}

func NewServiceUser(arch string) *ServiceUser {
	scope.Debug("NewServiceUser")

	return &ServiceUser{Arch: arch}
}

func (c *ServiceUser) IsUserExist(userName string) (bool, error) {
	user := entity.NewUserEntity(userName, "", "")

	exist, err := user.IsUserExist(userName)
	if err != nil {
		return exist, err
	}

	return exist, nil
}
