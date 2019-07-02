package apiserver

import (
	"fmt"
	Accounts "github.com/containers-ai/federatorai-api/apiserver/accounts"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Login() (string, error) {
	request := &Accounts.LoginRequest{
		Name: serverUsername,
		Password: serverPassword,
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
		return "", errors.New("failed to connect to api server")
	}
	defer conn.Close()

	client := Accounts.NewAccountsServiceClient(conn)
	response, err := client.Login(context.Background(), request)
	if err != nil {
		fmt.Print(err)
		return "", errors.New("failed to login api server")
	}

	return response.GetAccessToken(), nil
}
