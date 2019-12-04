package namespace

import (
	"context"
	"regexp"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	k8SUtils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
)

var scope = logUtil.RegisterScope("datahub_client_namespace", "namespace of datahub client", 0)

type NamespaceRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewNamespaceRepository return NamespaceRepository instance
func NewNamespaceRepository(conn *grpc.ClientConn, clusterUID string) *NamespaceRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &NamespaceRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

// CreateNamespaces creates namespaces to datahub
func (repo *NamespaceRepository) CreateNamespaces(arg interface{}) error {
	namespaces := []*datahub_resources.Namespace{}
	if nss, ok := arg.([]corev1.Namespace); ok {
		for _, ns := range nss {
			if !repo.IsNSExcluded(ns.GetName()) {
				namespaces = append(namespaces, &datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name:        ns.GetName(),
						ClusterName: repo.clusterUID,
					},
				})
			}
		}
	}
	if nss, ok := arg.([]*datahub_resources.Namespace); ok {
		for _, ns := range nss {
			if !repo.IsNSExcluded(ns.GetObjectMeta().GetName()) {
				namespaces = append(namespaces, ns)
			}
		}
	}

	req := datahub_resources.CreateNamespacesRequest{
		Namespaces: namespaces,
	}

	if resp, err := repo.datahubClient.CreateNamespaces(context.Background(), &req); err != nil {
		return errors.Wrap(err, "create namespaces to datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "create namespaces to datahub failed")
	}
	return nil
}

func (repo *NamespaceRepository) ListNamespaces() (
	[]*datahub_resources.Namespace, error) {
	req := datahub_resources.ListNamespacesRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}

	resp, err := repo.datahubClient.ListNamespaces(context.Background(), &req)
	if err != nil {
		return nil, errors.Wrap(err, "list namespaces from Datahub failed")
	} else if resp == nil {
		return nil, errors.Errorf("list namespaces from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list namespaces from Datahub failed")
	}
	return resp.Namespaces, nil
}

// DeleteNamespace delete namespaces from datahub
func (repo *NamespaceRepository) DeleteNamespaces(arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if nss, ok := arg.([]*corev1.Namespace); ok {
		for _, ns := range nss {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name:        ns.GetName(),
				ClusterName: repo.clusterUID,
			})
		}
	}
	if namespaces, ok := arg.([]*datahub_resources.Namespace); ok {
		for _, namespace := range namespaces {
			copyNamespace := *namespace
			objMeta = append(objMeta, copyNamespace.ObjectMeta)
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	req := datahub_resources.DeleteNamespacesRequest{
		ObjectMeta: objMeta,
	}

	resp, err := repo.datahubClient.DeleteNamespaces(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "delete namespace from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete namespace from Datahub failed")
	}
	return nil
}

func (repo *NamespaceRepository) Close() {
	repo.conn.Close()
}

func (repo *NamespaceRepository) IsNSExcluded(ns string) bool {
	req := &datahub_resources.ListApplicationsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{},
		},
	}
	resp, err := repo.datahubClient.ListApplications(context.Background(), req)
	if err != nil {
		scope.Errorf("namespace exclusion check error: %s", err.Error())
	} else {
		for _, app := range resp.GetApplications() {
			if app.GetObjectMeta().GetNamespace() == ns {
				return false
			}
		}
	}

	if ns == k8SUtils.GetRunningNamespace() {
		return true
	}

	excludeNamespaces := viper.GetStringSlice("namespace_exclusion.namespaces")
	excludeNSRegs := viper.GetStringSlice("namespace_exclusion.namespace_regs")
	for _, excludeNSReg := range excludeNSRegs {
		matched, _ := regexp.MatchString(excludeNSReg, ns)
		if matched {
			return true
		}
	}
	for _, excludeNamespace := range excludeNamespaces {
		if excludeNamespace == ns {
			return true
		}
	}
	return false
}
