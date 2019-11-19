package application

import (
	"context"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	"github.com/containers-ai/alameda/operator/datahub/client"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
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

	if resp, err := repo.datahubClient.CreateApplications(context.Background(), &req); err != nil {
		return errors.Errorf("create applications to datahub failed: %s", err.Error())
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "create applications to Datahub failed")
	}
	return nil
}

func (repo *ApplicationRepository) GetApplication(ctx context.Context, namespace, name string) (datahub_resources.Application, error) {
	req := datahub_resources.ListApplicationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   namespace,
				Name:        name,
				ClusterName: repo.clusterUID,
			},
		},
	}
	resp, err := repo.datahubClient.ListApplications(ctx, &req)
	if err != nil {
		return datahub_resources.Application{}, errors.Wrap(err, "list applications from Datahub failed")
	} else if resp == nil {
		return datahub_resources.Application{}, errors.New("list applications from Datahub failed: receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return datahub_resources.Application{}, errors.Wrap(err, "list applications from Datahub failed")
	}
	if len(resp.Applications) < 1 {
		return datahub_resources.Application{}, errors.New("not found")
	} else if resp.Applications[0] == nil {
		return datahub_resources.Application{}, errors.New("not found")
	} else if len(resp.Applications) > 1 {
		return datahub_resources.Application{}, errors.Errorf("length of applications from Datahub %d > 1", len(resp.Applications))
	}
	return *resp.Applications[0], nil
}

func (repo *ApplicationRepository) ListApplications() ([]*datahub_resources.Application, error) {
	req := datahub_resources.ListApplicationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}

	resp, err := repo.datahubClient.ListApplications(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrap(err, "list applications from datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list applications from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list applications from Datahub failed")
	}
	return resp.Applications, nil
}

// DeleteApplications delete applications from datahub
func (repo *ApplicationRepository) DeleteApplications(ctx context.Context, objectMetas []*datahub_resources.ObjectMeta) error {
	req := datahub_resources.DeleteApplicationsRequest{
		ObjectMeta: objectMetas,
	}
	if resp, err := repo.datahubClient.DeleteApplications(ctx, &req); err != nil {
		return errors.Wrap(err, "delete applications from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete applications from Datahub failed")
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
