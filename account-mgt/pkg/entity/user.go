package entity

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/objects"
	"github.com/containers-ai/alameda/apiserver/pkg/config"
)

// User : LDAP user struct
type User struct {
	Info   objects.UserInfo
	Config *config.Config
}

func NewUserEntity(name string, password string, token string, cfg *config.Config) *User {
	userInfo := objects.UserInfo{
		Name:     name,
		Password: password,
		Token:    token,
	}
	return &User{Info: userInfo, Config: cfg}
}
