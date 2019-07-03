package apiserver

import (
    "golang.org/x/net/context"
    "google.golang.org/grpc/metadata"
    "time"
)

var (
	serverAddr     = defaultAddress
	serverUsername = defaultUsername
	serverPassword = defaultPassword
	serverToken    = ""
)

func ServerInit(config Config) {
	serverAddr     = config.Address
	serverUsername = config.Username
	serverPassword = config.Password
}

func SetToken(token string) {
	serverToken = token
}

func GetToken(refresh bool, retry int) string {
	if refresh == true {
		for i := 0; i < retry; i++ {
			token, err := Login()
			if err == nil {
				serverToken = token
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	return serverToken
}

func NewContextWithCredential() context.Context {
	md := metadata.Pairs()
	md.Set("Username", serverUsername)
	md.Set("Password", serverPassword)
	md.Set("Token", serverToken)

    return metadata.NewOutgoingContext(context.Background(), md)
}
