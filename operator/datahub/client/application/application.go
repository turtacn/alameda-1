package application

import (
	"context"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

type ApplicationRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewApplicationRepository return ApplicationRepository instance
func NewApplicationRepository(conn *grpc.ClientConn, clusterUID string) *ApplicationRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &ApplicationRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

// CreateApplications creates applications to datahub
func (repo *ApplicationRepository) CreateApplications(arg interface{}) error {
	applications := []*datahub_resources.Application{}
	if apps, ok := arg.([]autoscalingv1alpha1.AlamedaScaler); ok {
		for _, app := range apps {
			applications = append(applications, &datahub_resources.Application{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        app.GetName(),
					Namespace:   app.GetNamespace(),
					ClusterName: repo.clusterUID,
				},
				AlamedaApplicationSpec: &datahub_resources.AlamedaApplicationSpec{
					ScalingTool: repo.getAlamedaScalerDatahubScalingType(app),
				},
			})
		}
	}
	if apps, ok := arg.([]*datahub_resources.Application); ok {
		applications = apps
	}

	req := datahub_resources.CreateApplicationsRequest{
		Applications: applications,
	}

	if reqRes, err := repo.datahubClient.CreateApplications(
		context.Background(), &req); err != nil {
		return errors.Errorf("create applications to datahub failed: %s", err.Error())
	} else if reqRes == nil {
		return errors.Errorf("create applications to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"create applications to datahub failed: receive statusCode: %d, message: %s",
			reqRes.Code, reqRes.Message)
	}
	return nil
}

func (repo *ApplicationRepository) ListApplications() (
	[]*datahub_resources.Application, error) {
	applications := []*datahub_resources.Application{}
	req := datahub_resources.ListApplicationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}
	if reqRes, err := repo.datahubClient.ListApplications(
		context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			return applications, errors.Errorf(
				"list applications from Datahub failed: %s", err.Error())
		}
		return applications, err
	} else {
		applications = reqRes.GetApplications()
	}
	return applications, nil
}

// DeleteApplication delete applications from datahub
func (repo *ApplicationRepository) DeleteApplications(
	arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if applications, ok := arg.([]*datahub_resources.Application); ok {
		for _, application := range applications {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        application.GetObjectMeta().GetName(),
				Namespace:   application.GetObjectMeta().GetNamespace(),
				ClusterName: repo.clusterUID,
			})
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	req := datahub_resources.DeleteApplicationsRequest{
		ObjectMeta: objMeta,
	}

	if resp, err := repo.datahubClient.DeleteApplications(
		context.Background(), &req); err != nil {
		return errors.Errorf("delete application from Datahub failed: %s",
			err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"delete application from Datahub failed: receive code: %d, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}

func (repo *ApplicationRepository) Close() {
	repo.conn.Close()
}

func (repo *ApplicationRepository) getAlamedaScalerDatahubScalingType(alamedaScaler autoscalingv1alpha1.AlamedaScaler) datahub_resources.ScalingTool {
	scalingType := datahub_resources.ScalingTool_SCALING_TOOL_UNDEFINED
	switch alamedaScaler.Spec.ScalingTool.Type {
	case autoscalingv1alpha1.ScalingToolTypeVPA:
		scalingType = datahub_resources.ScalingTool_VPA
	case autoscalingv1alpha1.ScalingToolTypeHPA:
		scalingType = datahub_resources.ScalingTool_HPA
	case autoscalingv1alpha1.ScalingToolTypeDefault:
		scalingType = datahub_resources.ScalingTool_NONE
	}
	return scalingType
}
