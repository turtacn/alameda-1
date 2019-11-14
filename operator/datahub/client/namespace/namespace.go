package namespace

import (
	"context"

	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

var scope = logUtil.RegisterScope("datahub_client_namespace", "namespace of datahub client", 0)

type NamespaceRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient
}

// NewNamespaceRepository return NamespaceRepository instance
func NewNamespaceRepository(conn *grpc.ClientConn) *NamespaceRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &NamespaceRepository{
		conn:          conn,
		datahubClient: datahubClient,
	}
}

// CreateNamespaces creates namespaces to datahub
func (repo *NamespaceRepository) CreateNamespaces(arg interface{}) error {
	namespaces := []*datahub_resources.Namespace{}
	if nss, ok := arg.([]corev1.Namespace); ok {
		for _, ns := range nss {
			if !repo.isNSExcluded(ns.GetName()) {
				namespaces = append(namespaces, &datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name: ns.GetName(),
					},
				})
			}
		}
	}
	if nss, ok := arg.([]*datahub_resources.Namespace); ok {
		for _, ns := range nss {
			if !repo.isNSExcluded(ns.GetObjectMeta().GetName()) {
				namespaces = append(namespaces, ns)
			}
		}
	}

	req := datahub_resources.CreateNamespacesRequest{
		Namespaces: namespaces,
	}

	if reqRes, err := repo.datahubClient.CreateNamespaces(
		context.Background(), &req); err != nil {
		return errors.Errorf("create namespaces to datahub failed: %s",
			err.Error())
	} else if reqRes == nil {
		return errors.Errorf("create namespaces to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"create namespaces to datahub failed: receive statusCode: %d, message: %s",
			reqRes.Code, reqRes.Message)
	}
	return nil
}

func (repo *NamespaceRepository) ListNamespaces() (
	[]*datahub_resources.Namespace, error) {
	namespaces := []*datahub_resources.Namespace{}
	req := datahub_resources.ListNamespacesRequest{}
	if reqRes, err := repo.datahubClient.ListNamespaces(
		context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			return namespaces, errors.Errorf(
				"list namespaces from Datahub failed: %s", err.Error())
		}
		return namespaces, err
	} else {
		namespaces = reqRes.GetNamespaces()
	}
	return namespaces, nil
}

// DeleteNamespace delete namespaces from datahub
func (repo *NamespaceRepository) DeleteNamespaces(arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if nss, ok := arg.([]*corev1.Namespace); ok {
		for _, ns := range nss {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name: ns.GetName(),
			})
		}
	}
	if namespaces, ok := arg.([]*datahub_resources.Namespace); ok {
		for _, namespace := range namespaces {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name: namespace.ObjectMeta.GetName(),
			})
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	req := datahub_resources.DeleteNamespacesRequest{
		ObjectMeta: objMeta,
	}

	if resp, err := repo.datahubClient.DeleteNamespaces(
		context.Background(), &req); err != nil {
		return errors.Errorf("delete namespace from Datahub failed: %s",
			err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf(
			"delete namespace from Datahub failed: receive code: %d, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}

func (repo *NamespaceRepository) Close() {
	repo.conn.Close()
}

func (repo *NamespaceRepository) isNSExcluded(ns string) bool {
	excludeNamespaces := viper.GetStringSlice("exclude_namespaces")
	for _, excludeNamespace := range excludeNamespaces {
		if excludeNamespace == ns {
			return true
		}
	}
	return false
}
