package entity

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/objects"
)

// User : LDAP user struct
type User struct {
	Info objects.UserInfo
}

func NewUserEntity(name string, domainName string, role string) *User {
	userInfo := objects.UserInfo{
		Name:       name,
		DomainName: domainName,
		Role:       role,
	}
	return &User{Info: userInfo}
}
