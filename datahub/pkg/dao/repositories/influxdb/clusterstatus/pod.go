package clusterstatus

import (
	//"fmt"
	"fmt"
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strings"
)

type PodRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewPodRepository(influxDBCfg *InternalInflux.Config) *PodRepository {
	return &PodRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *PodRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxCluster.PodTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

func (p *PodRepository) CreatePods(pods []*DaoClusterTypes.Pod) error {
	points := make([]*InfluxClient.Point, 0)

	for _, pod := range pods {
		entity := EntityInfluxCluster.PodEntity{
			Time:         InternalInflux.ZeroTime,
			Name:         pod.ObjectMeta.Name,
			Namespace:    pod.ObjectMeta.Namespace,
			NodeName:     pod.ObjectMeta.NodeName,
			ClusterName:  pod.ObjectMeta.ClusterName,
			Uid:          pod.ObjectMeta.Uid,
			CreateTime:   pod.CreateTime.GetSeconds(),
			ResourceLink: pod.ResourceLink,
			AppName:      pod.AppName,
			AppPartOf:    pod.AppPartOf,
		}
		if pod.TopController != nil {
			entity.TopControllerName = pod.TopController.ObjectMeta.Name
			entity.TopControllerKind = pod.TopController.Kind
			entity.TopControllerReplicas = pod.TopController.Replicas
		}
		if pod.Status != nil {
			entity.StatusPhase = pod.Status.Phase
			entity.StatusMessage = pod.Status.Message
			entity.StatusReason = pod.Status.Reason
		}
		if pod.AlamedaPodSpec != nil {
			entity.AlamedaSpecScalerName = pod.AlamedaPodSpec.AlamedaScaler.Name
			entity.AlamedaSpecScalerNamespace = pod.AlamedaPodSpec.AlamedaScaler.Namespace
			entity.AlamedaSpecScalerClusterName = pod.AlamedaPodSpec.AlamedaScaler.ClusterName
			entity.AlamedaSpecPolicy = pod.AlamedaPodSpec.Policy
			entity.AlamedaSpecUsedRecommendationID = pod.AlamedaPodSpec.UsedRecommendationId
			entity.AlamedaSpecScalingTool = pod.AlamedaPodSpec.ScalingTool
			if pod.AlamedaPodSpec.AlamedaScalerResources != nil {
				if value, exist := pod.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_CPU)]; exist {
					entity.AlamedaSpecResourceLimitCPU = value
				}
				if value, exist := pod.AlamedaPodSpec.AlamedaScalerResources.Limits[int32(ApiCommon.ResourceName_MEMORY)]; exist {
					entity.AlamedaSpecResourceLimitMemory = value
				}
				if value, exist := pod.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_CPU)]; exist {
					entity.AlamedaSpecResourceRequestCPU = value
				}
				if value, exist := pod.AlamedaPodSpec.AlamedaScalerResources.Requests[int32(ApiCommon.ResourceName_MEMORY)]; exist {
					entity.AlamedaSpecResourceRequestMemory = value
				}
			}
		}

		// Add to influx point list
		if pt, err := entity.BuildInfluxPoint(string(Pod)); err == nil {
			points = append(points, pt)
		} else {
			scope.Error(err.Error())
		}
	}

	// Batch write influxdb data points
	err := p.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "failed to batch write influxdb data points")
	}

	return nil
}

func (p *PodRepository) ListPods(request DaoClusterTypes.ListPodsRequest) ([]*DaoClusterTypes.Pod, error) {
	pods := make([]*DaoClusterTypes.Pod, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Pod,
		GroupByTags:    []string{string(EntityInfluxCluster.PodNamespace), string(EntityInfluxCluster.PodNodeName), string(EntityInfluxCluster.PodClusterName)},
	}

	// Build influx query command
	for _, objectMeta := range request.ObjectMeta {
		conditionList := make([]string, 0)

		metaCondition := p.genObjectMetaCondition(objectMeta, ApiResources.Kind(ApiResources.Kind_value[request.Kind]))
		if metaCondition != "" {
			conditionList = append(conditionList, metaCondition)
		}

		createCondition := p.genCreatePeriodCondition(request.QueryCondition)
		if createCondition != "" {
			conditionList = append(conditionList, createCondition)
		}

		condition := strings.Join(conditionList, " AND ")
		if condition != "" {
			condition = "(" + condition + ")"
		}
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	if len(request.ObjectMeta) == 0 {
		statement.AppendWhereClauseDirectly("AND", fmt.Sprintf(`("%s"='%s')`, EntityInfluxCluster.PodTopControllerKind, request.Kind))
		statement.AppendWhereClauseDirectly("AND", p.genCreatePeriodCondition(request.QueryCondition))
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make([]*DaoClusterTypes.Pod, 0), errors.Wrap(err, "failed to list pods")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				pod := DaoClusterTypes.NewPod()
				pod.Initialize(EntityInfluxCluster.NewPodEntity(row))
				pods = append(pods, pod)
			}
		}
	}

	return pods, nil
}

func (p *PodRepository) genObjectMetaCondition(objectMeta Metadata.ObjectMeta, kind ApiResources.Kind) string {
	conditions := make([]string, 0)

	switch kind {
	case ApiResources.Kind_POD:
		if objectMeta.Namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"=%s'`, EntityInfluxCluster.PodNamespace, objectMeta.Namespace))
		}
		if objectMeta.Name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodName, objectMeta.Name))
		}
	case ApiResources.Kind_DEPLOYMENT:
		conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerKind, ApiResources.Kind_name[int32(kind)]))
		if objectMeta.Namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodNamespace, objectMeta.Namespace))
		}
		if objectMeta.Name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerName, objectMeta.Name))
		}
	case ApiResources.Kind_DEPLOYMENTCONFIG:
		conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerKind, ApiResources.Kind_name[int32(kind)]))
		if objectMeta.Namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodNamespace, objectMeta.Namespace))
		}
		if objectMeta.Name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerName, objectMeta.Name))
		}
	case ApiResources.Kind_ALAMEDASCALER:
		if objectMeta.Namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodAlamedaSpecScalerNamespace, objectMeta.Namespace))
		}
		if objectMeta.Name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodAlamedaSpecScalerName, objectMeta.Name))
		}
	case ApiResources.Kind_STATEFULSET:
		conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerKind, ApiResources.Kind_name[int32(kind)]))
		if objectMeta.Namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodNamespace, objectMeta.Namespace))
		}
		if objectMeta.Name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s"='%s'`, EntityInfluxCluster.PodTopControllerName, objectMeta.Name))
		}
	default:
		scope.Errorf("no matching kind(%s) while building influx statement, skip it", kind)
		return ""
	}

	if len(conditions) > 0 {
		return strings.Join(conditions, " AND ")
	}

	return ""
}

func (p *PodRepository) genCreatePeriodCondition(query InternalCommon.QueryCondition) string {
	if query.StartTime != nil && query.EndTime != nil {
		return fmt.Sprintf("\"%s\">=%d AND \"%s\"<%d", EntityInfluxCluster.PodCreateTime, query.StartTime.Unix(), EntityInfluxCluster.PodCreateTime, query.EndTime.Unix())
	} else if query.StartTime != nil && query.EndTime == nil {
		return fmt.Sprintf("\"%s\">=%d", EntityInfluxCluster.PodCreateTime, query.StartTime.Unix())
	} else if query.StartTime == nil && query.EndTime != nil {
		return fmt.Sprintf("\"%s\"<%d", EntityInfluxCluster.PodCreateTime, query.EndTime.Unix())
	} else {
		return ""
	}
}
