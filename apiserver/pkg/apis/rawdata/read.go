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

func (c *ServiceRawdata) ReadRawdata(ctx context.Context, in *Rawdata.ReadRawdataRequest) (*Rawdata.ReadRawdataResponse, error) {
	scope.Debug("Request received from ReadRawdata grpc function")

	response := Rawdata.ReadRawdataResponse{}

	// Instance user entity
	user, password, token, err := Accounts.GetUserCredentialFromContext(ctx)
	if err != nil {
		scope.Errorf("get user credential error: %s", err.Error())
		response.Status = &status.Status{Code: int32(code.Code_INTERNAL), Message: "failed to get user credential"}
		return &response, nil
	}
	userInfo := entity.NewUserEntity(user, password, token, c.Config)

	// Do authentication
	err = userInfo.DoAuthentication()
	if err != nil {
		scope.Errorf("user authentication error: %s", err.Error())
		response.Status = &status.Status{
			Code:    int32(code.Code_UNAUTHENTICATED),
			Message: "invalid credential",
		}
		return &response, nil
	}

	// Instance datahub client
	conn, client, err := RepoDatahub.CreateClient(c.Config.Datahub.Address)
	if err != nil {
		response.Status = &status.Status{Code: int32(code.Code_INTERNAL), Message: err.Error()}
		return &response, nil
	}
	defer conn.Close()

	// Rebuild read rawdata request for datahub
	request := &Datahub.ReadRawdataRequest{}
	request.DatabaseType = in.GetDatabaseType()
	for _, query := range in.GetQueries() {
		request.Queries = append(request.Queries, query)
	}

	// Read rawdata from datahub
	if result, err := client.ReadRawdata(context.Background(), request); err != nil {
		scope.Errorf("apiserver ReadRawdata failed: %v", err)
		response.Status = &status.Status{Code: int32(code.Code_INTERNAL)}
	} else {
		response.Status = &status.Status{Code: int32(code.Code_OK)}
		response.Rawdata = result.Rawdata
	}

	return &response, nil
}
