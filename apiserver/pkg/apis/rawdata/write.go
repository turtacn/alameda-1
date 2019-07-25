package rawdata

import (
	"github.com/containers-ai/alameda/account-mgt/pkg/entity"
	Accounts "github.com/containers-ai/alameda/apiserver/pkg/apis/accounts"
	RepoDatahub "github.com/containers-ai/alameda/apiserver/pkg/repositories/datahub"
	Datahub "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	Rawdata "github.com/containers-ai/federatorai-api/apiserver/rawdata"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (c *ServiceRawdata) WriteRawdata(ctx context.Context, in *Rawdata.WriteRawdataRequest) (*status.Status, error) {
	scope.Debug("Request received from WriteRawdata grpc function")

	// Instance user entity
	user, password, token, err := Accounts.GetUserCredentialFromContext(ctx)
	if err != nil {
		scope.Errorf("get user credential error: %s", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: "failed to get user credential",
		}, nil
	}
	userInfo := entity.NewUserEntity(user, password, token, c.Config)

	// Do authentication
	err = userInfo.DoAuthentication()
	if err != nil {
		scope.Errorf("user authentication error: %s", err.Error())
		return &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: "invalid credential",
		}, nil
	}

	// Instance datahub client
	conn, client, err := RepoDatahub.CreateClient(c.Config.Datahub.Address)
	if err != nil {
		return &status.Status{Code: int32(code.Code_INTERNAL), Message: err.Error()}, nil
	}
	defer conn.Close()

	// Rebuild write rawdata request for datahub
	request := &Datahub.WriteRawdataRequest{}
	request.DatabaseType = in.GetDatabaseType()
	for _, rdata := range in.GetRawdata() {
		request.Rawdata = append(request.Rawdata, rdata)
	}

	// Write rawdata to datahub
	response, err := client.WriteRawdata(context.Background(), request)

	return response, err
}
