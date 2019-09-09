package node

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

var (
	scope = logUtil.RegisterScope("datahub node repository", "datahub node repository", 0)
)

// instance-typ --> size
var nodeLabels = `{      
	"beta.kubernetes.io/arch": "amd64",
	"beta.kubernetes.io/instance-type": "c4.xlarge",
	"beta.kubernetes.io/os": "linux",
	"failure-domain.beta.kubernetes.io/region": "us-west-2",
	"failure-domain.beta.kubernetes.io/zone": "us-west-2a",
	"kubernetes.io/arch": "amd64",
	"kubernetes.io/hostname": "ip-172-23-1-67.us-west-2.compute.internal",
	"kubernetes.io/os": "linux",
	"stackpoint.io/cluster_id": "23391",
	"stackpoint.io/instance_id": "netatt9dgn-worker-2",
	"stackpoint.io/node_group": "autoscaling-netatt9dgn-pool-1",
	"stackpoint.io/node_id": "91192",
	"stackpoint.io/node_pool": "Default-Worker-Pool",
	"stackpoint.io/private_ip": "172.23.1.67",
	"stackpoint.io/role": "worker",
	"stackpoint.io/size": "c4.xlarge"
}`

// providerID: aws:///us-west-2a/i-0769ec8570198bf4b --> <provider_raw>//<region>//<instance_id>

type nodeInfo struct {
	Name         string
	CreatedTime  int64
	Namespace    string
	Kind         string
	Role         string
	Region       string
	Zone         string
	Size         string
	InstanceType string
	OS           string
	Provider     string
	InstanceID   string
	StorageSize  int64
	CPUCores     int64
	MemoryBytes  int64
}

// newNode creates node from k8s node
func newNode(k8sNode corev1.Node) (nodeInfo, error) {
	node := nodeInfo{Name: k8sNode.Name, Namespace: k8sNode.Namespace, Kind: k8sNode.Kind}
	rf := reflect.TypeOf(node)
	rv := reflect.ValueOf(&node).Elem()
	for i := 0; i < rf.NumField(); i++ {
		key := rf.Field(i).Name
		// parse node label information
		for labelKey, labelV := range k8sNode.Labels {
			if strings.Contains(labelKey, "stackpoint.") && strings.Contains(labelKey, "stackpoint.io/role") == false {
				continue
			}
			value := parseKeyValue(labelKey, key, labelV)
			if len(value) > 0 {
				rValue := rv.FieldByName(strings.Title(key))
				rValue.SetString(string(labelV))
				break
			}
		}
		switch key {
		case "StorageSize":
			node.StorageSize = k8sNode.Status.Capacity.StorageEphemeral().Value()
		}
	}

	if len(k8sNode.Spec.ProviderID) > 0 {
		provider, _, instanceID := parseProviderID(k8sNode.Spec.ProviderID)
		node.Provider = provider
		node.InstanceID = instanceID
	}

	// Below ard original convert logic
	node.CreatedTime = k8sNode.ObjectMeta.GetCreationTimestamp().Unix()

	cpuCores, ok := k8sNode.Status.Capacity.Cpu().AsInt64()
	if !ok {
		return nodeInfo{}, errors.Errorf("cannot convert cpu capacity from k8s Node")
	}
	node.CPUCores = cpuCores

	memoryBytes, ok := k8sNode.Status.Capacity.Memory().AsInt64()
	if !ok {
		return nodeInfo{}, errors.Errorf("cannot convert memory capacity from k8s Node")
	}
	node.MemoryBytes = memoryBytes
 

	return node, nil
}

func (n nodeInfo) alamedaNode() datahub_v1alpha1.Node {

	node := datahub_v1alpha1.Node{
		Name: n.Name,
		Capacity: &datahub_v1alpha1.Capacity{
			CpuCores:    n.CPUCores,
			MemoryBytes: n.MemoryBytes,
		},
		StartTime: &timestamp.Timestamp{
			Seconds: n.CreatedTime,
		},
		Provider: &datahub_v1alpha1.Provider{
			Provider:     n.Provider,
			InstanceType: n.InstanceType,
			Region:       n.Region,
			Zone:         n.Zone,
			Os:           n.OS,
			Role:         n.Role,
			InstanceId:   n.InstanceID,
			StorageSize:  n.StorageSize,
		},
	}

	return node
}

func parseKeyValue(strParse string, key string, value string) string {
	pattern, err := regexp.Compile(strings.ToLower(fmt.Sprintf("/%s$", key)))
	if err != nil {
		return ""
	}
	if len(pattern.FindString(strings.Replace(strParse, "-", "", -1))) > 0 {
		return value
	}
	return ""
}

func parseProviderID(providerID string) (string, string, string) {
	var provider string
	var region string
	var instanceID string
	rex, err := regexp.Compile("([^\\:/]+)")
	if err != nil {
		fmt.Println(err)
		return "", "", ""
	}
	res := rex.FindAllString(providerID, -1)
	if res == nil || len(res) == 0 {
		return "", "", ""
	}
	for i := 0; i < len(res) && i < 3; i++ {
		switch i {
		case 0:
			provider = res[i]
		case 1:
			region = res[i]
		case 2:
			instanceID = res[i]
		}
	}
	return provider, region, instanceID
}

// AlamedaNodeRepository creates predicted node to datahub
type AlamedaNodeRepository struct{}

// NewAlamedaNodeRepository return AlamedaNodeRepository instance
func NewAlamedaNodeRepository() *AlamedaNodeRepository {
	return &AlamedaNodeRepository{}
}

// CreateAlamedaNode creates predicted node to datahub
func (repo *AlamedaNodeRepository) CreateAlamedaNode(nodes []*corev1.Node) error {
	retries := 3
	for retry := 1; retry <= retries; retry++ {
		err := repo.createAlamedaNode(nodes)
		if err == nil {
			break
		}
		scope.Debugf("Create Alameda node failed. (%v try)", retry)
		if retry == retries {
			return err
		}
	}
	return nil
}

func (repo *AlamedaNodeRepository) createAlamedaNode(nodes []*corev1.Node) error {
	alamedaNodes := []*datahub_v1alpha1.Node{}
	for _, node := range nodes {
		n, err := newNode(*node)
		if err != nil {
			scope.Errorf("Create nodeInfo failed, skip creating node (%s) to Datahub, error message: %s", node.GetName(), err.Error())
			continue
		}
		alamedaNode := n.alamedaNode()
		alamedaNodes = append(alamedaNodes, &alamedaNode)
	}
	req := datahub_v1alpha1.CreateAlamedaNodesRequest{
		AlamedaNodes: alamedaNodes,
	}
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if reqRes, err := datahubServiceClnt.CreateAlamedaNodes(context.Background(), &req); err != nil {
		scope.Error(fmt.Sprintf("Create nodes to datahub failed: %s", err.Error()))
		return err
	} else if reqRes == nil {
		return errors.Errorf("Create nodes to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf("Create nodes to datahub failed: receive statusCode: %d, message: %s", reqRes.Code, reqRes.Message)
	}
	return nil
}

// DeleteAlamedaNodes delete predicted node from datahub
func (repo *AlamedaNodeRepository) DeleteAlamedaNodes(nodes []*corev1.Node) error {

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		return errors.Wrapf(err, "delete node from Datahub failed: %s", err.Error())
	}

	alamedaNodes := []*datahub_v1alpha1.Node{}
	for _, node := range nodes {
		alamedaNodes = append(alamedaNodes, &datahub_v1alpha1.Node{
			Name: node.GetName(),
		})
	}
	req := datahub_v1alpha1.DeleteAlamedaNodesRequest{
		AlamedaNodes: alamedaNodes,
	}

	aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if resp, err := aiServiceClnt.DeleteAlamedaNodes(context.Background(), &req); err != nil {
		return errors.Wrapf(err, "delete node from Datahub failed: %s", err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("delete node from Datahub failed: receive code: %d, message: %s", resp.Code, resp.Message)
	}
	return nil
}

// ListAlamedaNodes lists nodes to datahub
func (repo *AlamedaNodeRepository) ListAlamedaNodes() ([]*datahub_v1alpha1.Node, error) {
	retries := 3
	alamNodes := []*datahub_v1alpha1.Node{}
	for retry := 1; retry <= retries; retry++ {
		nodes, err := repo.listAlamedaNodes()
		if err == nil {
			alamNodes = nodes
			break
		}
		scope.Debugf("List alameda nodes failed. (%v try)", retry)
		if retry == retries {
			return nil, err
		}
	}
	return alamNodes, nil
}

func (repo *AlamedaNodeRepository) listAlamedaNodes() ([]*datahub_v1alpha1.Node, error) {
	alamNodes := []*datahub_v1alpha1.Node{}
	req := datahub_v1alpha1.ListAlamedaNodesRequest{}
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		scope.Error(err.Error())
		return nil, err
	}

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if reqRes, err := datahubServiceClnt.ListAlamedaNodes(context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			scope.Error(reqRes.Status.GetMessage())
		}
		return alamNodes, err
	} else {
		alamNodes = reqRes.GetNodes()
	}
	return alamNodes, nil
}
