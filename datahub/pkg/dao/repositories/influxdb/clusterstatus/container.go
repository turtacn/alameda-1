package clusterstatus

import (
	//"fmt"
	EntityInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	//FormatConvert "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalInfluxModels "github.com/containers-ai/alameda/internal/pkg/database/influxdb/models"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	//ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	//"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	//"strconv"
	//"strings"
	"strings"
)

type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

func NewContainerRepository(influxDBCfg *InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

func (p *ContainerRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxCluster.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

func (p *ContainerRepository) CreateContainers(containers map[string][]*DaoClusterTypes.Container) error {
	points := make([]*InfluxClient.Point, 0)

	for _, cnts := range containers {
		for _, cnt := range cnts {
			entity := EntityInfluxCluster.ContainerEntity{
				Time:        InternalInflux.ZeroTime,
				Name:        cnt.Name,
				PodName:     cnt.PodName,
				Namespace:   cnt.Namespace,
				NodeName:    cnt.NodeName,
				ClusterName: cnt.ClusterName,
				Uid:         cnt.Uid,
			}
			if cnt.Resources != nil {
				if value, exist := cnt.Resources.Limits[int32(ApiCommon.ResourceName_CPU)]; exist {
					entity.ResourceLimitCPU = value
				}
				if value, exist := cnt.Resources.Limits[int32(ApiCommon.ResourceName_MEMORY)]; exist {
					entity.ResourceLimitMemory = value
				}
				if value, exist := cnt.Resources.Requests[int32(ApiCommon.ResourceName_CPU)]; exist {
					entity.ResourceRequestCPU = value
				}
				if value, exist := cnt.Resources.Requests[int32(ApiCommon.ResourceName_MEMORY)]; exist {
					entity.ResourceRequestMemory = value
				}
			}
			if cnt.Status != nil {
				if cnt.Status.State != nil {
					if cnt.Status.State.Waiting != nil {
						entity.StatusWaitingReason = cnt.Status.State.Waiting.Reason
						entity.StatusWaitingMessage = cnt.Status.State.Waiting.Message
					}
					if cnt.Status.State.Running != nil {
						entity.StatusRunningStartedAt = cnt.Status.State.Running.StartedAt.GetSeconds()
					}
					if cnt.Status.State.Terminated != nil {
						entity.StatusTerminatedExitCode = cnt.Status.State.Terminated.ExitCode
						entity.StatusTerminatedReason = cnt.Status.State.Terminated.Reason
						entity.StatusTerminatedMessage = cnt.Status.State.Terminated.Message
						entity.StatusTerminatedStartedAt = cnt.Status.State.Terminated.StartedAt.GetSeconds()
						entity.StatusTerminatedFinishedAt = cnt.Status.State.Terminated.FinishedAt.GetSeconds()
					}

				}
				if cnt.Status.LastTerminationState != nil {
					if cnt.Status.LastTerminationState.Waiting != nil {
						entity.LastTerminationWaitingReason = cnt.Status.LastTerminationState.Waiting.Reason
						entity.LastTerminationWaitingMessage = cnt.Status.LastTerminationState.Waiting.Message
					}
					if cnt.Status.LastTerminationState.Running != nil {
						entity.LastTerminationRunningStartedAt = cnt.Status.LastTerminationState.Running.StartedAt.GetSeconds()
					}
					if cnt.Status.LastTerminationState.Terminated != nil {
						entity.LastTerminationTerminatedExitCode = cnt.Status.LastTerminationState.Terminated.ExitCode
						entity.LastTerminationTerminatedReason = cnt.Status.LastTerminationState.Terminated.Reason
						entity.LastTerminationTerminatedMessage = cnt.Status.LastTerminationState.Terminated.Message
						entity.LastTerminationTerminatedStartedAt = cnt.Status.LastTerminationState.Terminated.StartedAt.GetSeconds()
						entity.LastTerminationTerminatedFinishedAt = cnt.Status.LastTerminationState.Terminated.FinishedAt.GetSeconds()
					}
				}
			}
			entity.RestartCount = cnt.Status.RestartCount

			// Add to influx point list
			if pt, err := entity.BuildInfluxPoint(string(Container)); err == nil {
				points = append(points, pt)
			} else {
				scope.Error(err.Error())
			}
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

func (p *ContainerRepository) ListContainers(request DaoClusterTypes.ListContainersRequest) (map[string][]*DaoClusterTypes.Container, error) {
	containerMap := make(map[string][]*DaoClusterTypes.Container, 0)

	statement := InternalInflux.Statement{
		QueryCondition: &request.QueryCondition,
		Measurement:    Container,
		GroupByTags:    []string{string(EntityInfluxCluster.ContainerPodName), string(EntityInfluxCluster.ContainerNamespace), string(EntityInfluxCluster.ContainerNodeName), string(EntityInfluxCluster.ContainerClusterName)},
	}

	// Build influx query command
	for _, containerMeta := range request.ContainerObjectMeta {
		keyList := containerMeta.ObjectMeta.GenerateKeyList()
		keyList = append(keyList, string(EntityInfluxCluster.ContainerPodName))

		valueList := containerMeta.ObjectMeta.GenerateValueList()
		valueList = append(valueList, containerMeta.PodName)

		condition := statement.GenerateCondition(keyList, valueList, "AND")
		statement.AppendWhereClauseDirectly("OR", condition)
	}
	statement.SetOrderClauseFromQueryCondition()
	statement.SetLimitClauseFromQueryCondition()
	cmd := statement.BuildQueryCmd()

	response, err := p.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return make(map[string][]*DaoClusterTypes.Container, 0), errors.Wrap(err, "failed to list containers")
	}

	results := InternalInfluxModels.NewInfluxResults(response)
	for _, result := range results {
		for i := 0; i < result.GetGroupNum(); i++ {
			group := result.GetGroup(i)
			row := group.GetRow(0)
			clusterNamespacePodName := p.ClusterNamespacePodName(row)
			containerMap[clusterNamespacePodName] = make([]*DaoClusterTypes.Container, 0)
			for j := 0; j < group.GetRowNum(); j++ {
				row := group.GetRow(j)
				container := DaoClusterTypes.NewContainer()
				container.Initialize(EntityInfluxCluster.NewContainerEntity(row))
				containerMap[clusterNamespacePodName] = append(containerMap[clusterNamespacePodName], container)
			}
		}
	}

	return containerMap, nil
}

func (p *ContainerRepository) ClusterNamespacePodName(values map[string]string) string {
	valueList := make([]string, 0)

	if value, ok := values[string(EntityInfluxCluster.ContainerClusterName)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}
	if value, ok := values[string(EntityInfluxCluster.ContainerNamespace)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}
	if value, ok := values[string(EntityInfluxCluster.ContainerPodName)]; ok {
		if value != "" {
			valueList = append(valueList, value)
		}
	}

	if len(valueList) > 0 {
		return strings.Join(valueList, "/")
	}

	return ""
}

// CreateContainers add containers information container measurement
/*func (p *ContainerRepository) CreateContainersOrig(pods []*ApiResources.Pod) error {
	points := make([]*InfluxClient.Point, 0)

	// Do delete containers before creating them
	err := containerRepository.DeleteContainers(pods)
	if err != nil {
		scope.Error("failed to delete container in influxdb when creating containers")
		return errors.Wrap(err, "failed to create containers to influxdb")
	}

	// Generate influxdb points
	for _, pod := range pods {
		containerEntities, err := buildContainerEntitiesFromDatahubPod(pod)
		if err != nil {
			return errors.Wrap(err, "failed to create containers")
		}
		for _, containerEntity := range containerEntities {
			p, err := containerEntity.InfluxDBPoint(string(Container))
			if err != nil {
				return errors.Wrap(err, "failed to create containers")
			}
			points = append(points, p)
		}
	}

	// Write points to influxdb
	err = containerRepository.influxDB.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(RepoInflux.ClusterStatus),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create containers to influxdb")
	}

	return nil
}*/

// ListAlamedaContainers list predicted containers have relation with arguments
/*

// DeleteContainers set containers' field is_deleted to true into container measurement
func (p *ContainerRepository) DeleteContainers(pods []*ApiResources.Pod) error {
	for _, pod := range pods {
		if pod.GetObjectMeta() == nil || pod.GetAlamedaPodSpec().GetAlamedaScaler() == nil {
			continue
		}
		podNS := pod.GetObjectMeta().GetNamespace()
		podName := pod.GetObjectMeta().GetName()
		alaScalerNS := pod.GetAlamedaPodSpec().GetAlamedaScaler().GetNamespace()
		alaScalerName := pod.GetAlamedaPodSpec().GetAlamedaScaler().GetName()
		cmd := fmt.Sprintf("DROP SERIES FROM %s WHERE \"%s\"='%s' AND \"%s\"='%s' AND \"%s\"='%s' AND \"%s\"='%s'", Container,
			EntityInfluxClusterStatus.ContainerNamespace, podNS, EntityInfluxClusterStatus.ContainerPodName, podName,
			EntityInfluxClusterStatus.ContainerAlamedaScalerNamespace, alaScalerNS, EntityInfluxClusterStatus.ContainerAlamedaScalerName, alaScalerName)
		scope.Debugf("DeleteContainers command: %s", cmd)
		_, err := containerRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
		if err != nil {
			scope.Errorf(err.Error())
		}
	}
	return nil
}*/

// ListPodsContainers list containers information container measurement
/*func (p *ContainerRepository) ListPodsContainers(pods []*ApiResources.Pod) ([]*EntityInfluxClusterStatus.ContainerEntity, error) {

	var (
		cmd                 = ""
		cmdSelectString     = ""
		cmdTagsFilterString = ""
		containerEntities   = make([]*EntityInfluxClusterStatus.ContainerEntity, 0)
	)

	if len(pods) == 0 {
		return containerEntities, nil
	}

	cmdSelectString = fmt.Sprintf(`select * from "%s" `, Container)
	for _, pod := range pods {

		var (
			namespace = ""
			podName   = ""
		)

		if pod.GetObjectMeta() != nil {
			namespace = pod.GetObjectMeta().GetNamespace()
			podName = pod.GetObjectMeta().GetName()
		}

		cmdTagsFilterString += fmt.Sprintf(`("%s" = '%s' and "%s" = '%s') or `,
			EntityInfluxClusterStatus.ContainerNamespace, namespace,
			EntityInfluxClusterStatus.ContainerPodName, podName,
		)
	}
	cmdTagsFilterString = strings.TrimSuffix(cmdTagsFilterString, "or ")

	cmd = fmt.Sprintf("%s where %s", cmdSelectString, cmdTagsFilterString)
	results, err := containerRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus))
	if err != nil {
		return containerEntities, errors.Wrap(err, "list pod containers failed")
	}

	rows := InternalInflux.PackMap(results)
	for _, row := range rows {
		for _, data := range row.Data {
			entity := EntityInfluxClusterStatus.NewContainerEntityFromMap(data)
			containerEntities = append(containerEntities, &entity)
		}
	}

	return containerEntities, nil
}*/

/*
func buildContainerEntitiesFromDatahubPod(pod *ApiResources.Pod) ([]*EntityInfluxClusterStatus.ContainerEntity, error) {

	var (
		namespace                                 *string
		podName                                   *string
		podPhase                                  *string
		podMessage                                *string
		podReason                                 *string
		alamedaScalerNamespace                    *string
		alamedaScalerName                         *string
		nodeName                                  *string
		name                                      *string
		statusWaitingReason                       *string
		statusWaitingMessage                      *string
		statusRunningStartedAt                    *int64
		statusTerminatedExitCode                  *int32
		statusTerminatedReason                    *string
		statusTerminatedMessage                   *string
		statusTerminatedStartedAt                 *int64
		statusTerminatedFinishedAt                *int64
		lastTerminationStatusWaitingReason        *string
		lastTerminationStatusWaitingMessage       *string
		lastTerminationStatusRunningStartedAt     *int64
		lastTerminationStatusTerminatedExitCode   *int32
		lastTerminationStatusTerminatedReason     *string
		lastTerminationStatusTerminatedMessage    *string
		lastTerminationStatusTerminatedStartedAt  *int64
		lastTerminationStatusTerminatedFinishedAt *int64
		restartCount                              *int32
		resourceRequestCPU                        *float64
		resourceRequestMemory                     *int64
		resourceLimitCPU                          *float64
		resourceLimitMemory                       *int64
		policy                                    *string
		podCreatedTime                            *int64
		resourceLink                              *string
		topControllerName                         *string
		topControllerKind                         *string
		topControllerReplicas                     *int32
		usedRecommendationID                      *string
		alamedaScalerResourceLimitCPU             *float64
		alamedaScalerResourceLimitMemory          *float64
		alamedaScalerResourceRequestCPU           *float64
		alamedaScalerResourceRequestMemory        *float64
	)

	nodeName = &pod.ObjectMeta.NodeName
	resourceLink = &pod.ResourceLink
	usedRecommendationID = &pod.AlamedaPodSpec.UsedRecommendationId

	if pod.ObjectMeta != nil {
		namespace = &pod.ObjectMeta.Namespace
		podName = &pod.ObjectMeta.Name
	}
	if pod.AlamedaPodSpec.AlamedaScaler != nil {
		alamedaScalerNamespace = &pod.AlamedaPodSpec.AlamedaScaler.Namespace
		alamedaScalerName = &pod.AlamedaPodSpec.AlamedaScaler.Name
	}
	if pod.StartTime != nil {
		startTime := pod.StartTime.GetSeconds()
		podCreatedTime = &startTime
	}
	if pod.TopController != nil {
		if pod.TopController.ObjectMeta != nil {
			topControllerName = &pod.TopController.ObjectMeta.Name
		}
		if k, exist := FormatConvert.KindDisp[pod.TopController.Kind]; exist {
			topControllerKind = &k
		}
		topControllerReplicas = &pod.TopController.Replicas
	}
	if p, exist := FormatConvert.RecommendationPolicyDisp[pod.GetAlamedaPodSpec().GetPolicy()]; exist {
		policy = &p
	}
	if pod.Status != nil {
		if val, ok := ApiResources.PodPhase_name[int32(pod.Status.Phase)]; ok {
			podPhase = &val
		} else {
			val = ApiResources.PodPhase_name[int32(ApiResources.PodPhase_UNKNOWN)]
			podPhase = &val
		}
		podMessage = &pod.Status.Message
		podReason = &pod.Status.Reason
	}

	appName := pod.GetAppName()
	appPartOf := pod.GetAppPartOf()
	// TODO: transform enableVPA and enableHAP to into ScalingTool
	enableVPA := true
	enableHPA := false

	containerEntities := make([]*EntityInfluxClusterStatus.ContainerEntity, 0)
	for _, datahubContainer := range pod.Containers {

		name = &datahubContainer.Name

		resourceLimitCPU = nil
		resourceLimitMemory = nil
		resourceRequestCPU = nil
		resourceRequestMemory = nil
		statusWaitingReason = nil
		statusWaitingMessage = nil
		statusRunningStartedAt = nil
		statusTerminatedExitCode = nil
		statusTerminatedReason = nil
		statusTerminatedMessage = nil
		statusTerminatedStartedAt = nil
		statusTerminatedFinishedAt = nil
		lastTerminationStatusWaitingReason = nil
		lastTerminationStatusWaitingMessage = nil
		lastTerminationStatusRunningStartedAt = nil
		lastTerminationStatusTerminatedExitCode = nil
		lastTerminationStatusTerminatedReason = nil
		lastTerminationStatusTerminatedMessage = nil
		lastTerminationStatusTerminatedStartedAt = nil
		lastTerminationStatusTerminatedFinishedAt = nil
		restartCount = nil

		if datahubContainer.GetResources() != nil {
			resourceRequirements := datahubContainer.GetResources()
			if resourceRequirements.GetLimits() != nil {
				for resourceName, value := range datahubContainer.GetResources().GetLimits() {
					switch ApiCommon.ResourceName(resourceName) {
					case ApiCommon.ResourceName_CPU:
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						resourceLimitCPU = &val
					case ApiCommon.ResourceName_MEMORY:
						val, err := strconv.ParseInt(value, 10, 64)
						if err != nil {
							scope.Warnf("convert string: %s to int64 failed, skip assigning value, err: %s", value, err.Error())
						}
						resourceLimitMemory = &val
					default:
						scope.Warnf("no mapping resource type for Datahub.ResourceName: %s, skip assigning value", ApiCommon.ResourceName_name[int32(resourceName)])
					}
				}
			}
			if resourceRequirements.GetRequests() != nil {
				for resourceName, value := range datahubContainer.GetResources().GetRequests() {
					switch ApiCommon.ResourceName(resourceName) {
					case ApiCommon.ResourceName_CPU:
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						resourceRequestCPU = &val
					case ApiCommon.ResourceName_MEMORY:
						val, err := strconv.ParseInt(value, 10, 64)
						if err != nil {
							scope.Warnf("convert string: %s to int64 failed, skip assigning value, err: %s", value, err.Error())
						}
						resourceRequestMemory = &val
					default:
						scope.Warnf("no mapping resource type for Datahub.ResourceName: %s, skip assigning value", ApiCommon.ResourceName_name[int32(resourceName)])
					}
				}
			}
		}
		if datahubContainer.GetStatus() != nil {
			containerStatus := datahubContainer.GetStatus()
			if containerStatus.GetState() != nil {
				state := containerStatus.GetState()
				if state.GetWaiting() != nil {
					statusWaitingReason = &state.GetWaiting().Reason
					statusWaitingMessage = &state.GetWaiting().Message
				}
				if state.GetRunning() != nil {
					statusRunningStartedAt = &state.GetRunning().GetStartedAt().Seconds
				}
				if state.GetTerminated() != nil {
					statusTerminatedExitCode = &state.GetTerminated().ExitCode
					statusTerminatedReason = &state.GetTerminated().Reason
					statusTerminatedMessage = &state.GetTerminated().Message
					statusTerminatedStartedAt = &state.GetTerminated().GetStartedAt().Seconds
					statusTerminatedFinishedAt = &state.GetTerminated().GetFinishedAt().Seconds
				}
			}
			if containerStatus.GetLastTerminationState() != nil {
				state := containerStatus.GetLastTerminationState()
				if state.GetWaiting() != nil {
					lastTerminationStatusWaitingReason = &state.GetWaiting().Reason
					lastTerminationStatusWaitingMessage = &state.GetWaiting().Message
				}
				if state.GetRunning() != nil {
					lastTerminationStatusRunningStartedAt = &state.GetRunning().GetStartedAt().Seconds
				}
				if state.GetTerminated() != nil {
					lastTerminationStatusTerminatedExitCode = &state.GetTerminated().ExitCode
					lastTerminationStatusTerminatedReason = &state.GetTerminated().Reason
					lastTerminationStatusTerminatedMessage = &state.GetTerminated().Message
					lastTerminationStatusTerminatedStartedAt = &state.GetTerminated().GetStartedAt().Seconds
					lastTerminationStatusTerminatedFinishedAt = &state.GetTerminated().GetFinishedAt().Seconds
				}
			}
			restartCount = &containerStatus.RestartCount
		}
		if pod.GetAlamedaPodSpec().GetAlamedaScalerResources() != nil {
			resourceRequirements := pod.GetAlamedaPodSpec().GetAlamedaScalerResources()
			if resourceRequirements.GetLimits() != nil {
				for resourceName, value := range resourceRequirements.GetLimits() {
					switch resourceName {
					case int32(ApiCommon.ResourceName_CPU):
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						alamedaScalerResourceLimitCPU = &val
					case int32(ApiCommon.ResourceName_MEMORY):
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						alamedaScalerResourceLimitMemory = &val
					default:
						scope.Warnf("no mapping resource name for Datahub.ResourceName: %d, skip assigning value", resourceName)
					}
				}
			}
			if resourceRequirements.GetRequests() != nil {
				for resourceName, value := range resourceRequirements.GetRequests() {
					switch resourceName {
					case int32(ApiCommon.ResourceName_CPU):
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						alamedaScalerResourceRequestCPU = &val
					case int32(ApiCommon.ResourceName_MEMORY):
						val, err := strconv.ParseFloat(value, 64)
						if err != nil {
							scope.Warnf("convert string: %s to float64 failed, skip assigning value, err: %s", value, err.Error())
						}
						alamedaScalerResourceRequestMemory = &val
					default:
						scope.Warnf("no mapping resource name for Datahub.ResourceName: %d, skip assigning value", resourceName)
					}
				}
			}
		}

		containerEntity := &EntityInfluxClusterStatus.ContainerEntity{
			Time:                                      InternalInflux.ZeroTime,
			Namespace:                                 namespace,
			PodName:                                   podName,
			PodPhase:                                  podPhase,
			PodMessage:                                podMessage,
			PodReason:                                 podReason,
			AlamedaScalerNamespace:                    alamedaScalerNamespace,
			AlamedaScalerName:                         alamedaScalerName,
			NodeName:                                  nodeName,
			Name:                                      name,
			StatusWaitingReason:                       statusWaitingReason,
			StatusWaitingMessage:                      statusWaitingMessage,
			StatusRunningStartedAt:                    statusRunningStartedAt,
			StatusTerminatedExitCode:                  statusTerminatedExitCode,
			StatusTerminatedReason:                    statusTerminatedReason,
			StatusTerminatedMessage:                   statusTerminatedMessage,
			StatusTerminatedStartedAt:                 statusTerminatedStartedAt,
			StatusTerminatedFinishedAt:                statusTerminatedFinishedAt,
			LastTerminationStatusWaitingReason:        lastTerminationStatusWaitingReason,
			LastTerminationStatusWaitingMessage:       lastTerminationStatusWaitingMessage,
			LastTerminationStatusRunningStartedAt:     lastTerminationStatusRunningStartedAt,
			LastTerminationStatusTerminatedExitCode:   lastTerminationStatusTerminatedExitCode,
			LastTerminationStatusTerminatedReason:     lastTerminationStatusTerminatedReason,
			LastTerminationStatusTerminatedMessage:    lastTerminationStatusTerminatedMessage,
			LastTerminationStatusTerminatedStartedAt:  lastTerminationStatusTerminatedStartedAt,
			LastTerminationStatusTerminatedFinishedAt: lastTerminationStatusTerminatedFinishedAt,
			RestartCount:                              restartCount,
			ResourceRequestCPU:                        resourceRequestCPU,
			ResourceRequestMemory:                     resourceRequestMemory,
			ResourceLimitCPU:                          resourceLimitCPU,
			ResourceLimitMemory:                       resourceLimitMemory,
			Policy:                                    policy,
			PodCreatedTime:                            podCreatedTime,
			ResourceLink:                              resourceLink,
			TopControllerName:                         topControllerName,
			TopControllerKind:                         topControllerKind,
			TpoControllerReplicas:                     topControllerReplicas,
			UsedRecommendationID:                      usedRecommendationID,
			AppName:                                   &appName,
			AppPartOf:                                 &appPartOf,
			EnableHPA:                                 &enableHPA,
			EnableVPA:                                 &enableVPA,
			AlamedaScalerResourceLimitCPU:             alamedaScalerResourceLimitCPU,
			AlamedaScalerResourceLimitMemory:          alamedaScalerResourceLimitMemory,
			AlamedaScalerResourceRequestCPU:           alamedaScalerResourceRequestCPU,
			AlamedaScalerResourceRequestMemory:        alamedaScalerResourceRequestMemory,
		}
		containerEntities = append(containerEntities, containerEntity)
	}
	return containerEntities, nil
}
*/

/*
func buildDatahubPodsFromContainerEntities(containerEntities []*EntityInfluxClusterStatus.ContainerEntity) []*ApiResources.Pod {

	datahubPods := make([]*ApiResources.Pod, 0)
	datahubPodMap := make(map[string]*ApiResources.Pod)

	for _, containerEntity := range containerEntities {

		podID := getDatahubPodIDString(containerEntity)

		var datahubPod *ApiResources.Pod
		datahubPod, exist := datahubPodMap[podID]
		if !exist {

			var (
				podName                string
				podPhase               string
				podMessage             string
				podReason              string
				namespace              string
				resourceLink           string
				alamedaScalerNamespace string
				alamedaScalerName      string
				nodeName               string
				podCreatedTime         int64
				policy                 string
				topControllerNamespace string
				topControllerName      string
				topControllerKind      string
				topControllerReplicas  int32
				usedRecommendationID   string
				appName                string
				appPartOf              string
				// TODO: add new member scalingTool
				// enableHPA                          bool
				// enableVPA                          bool
				alamedaScalerResourceLimitCPU      string
				alamedaScalerResourceLimitMemory   string
				alamedaScalerResourceRequestCPU    string
				alamedaScalerResourceRequestMemory string
			)

			if containerEntity.PodName != nil {
				podName = *containerEntity.PodName
			}
			if containerEntity.PodPhase != nil {
				podPhase = *containerEntity.PodPhase
			}
			if containerEntity.PodMessage != nil {
				podMessage = *containerEntity.PodMessage
			}
			if containerEntity.PodReason != nil {
				podReason = *containerEntity.PodReason
			}
			if containerEntity.Namespace != nil {
				namespace = *containerEntity.Namespace
			}
			if containerEntity.ResourceLink != nil {
				resourceLink = *containerEntity.ResourceLink
			}
			if containerEntity.AlamedaScalerNamespace != nil {
				alamedaScalerNamespace = *containerEntity.AlamedaScalerNamespace
			}
			if containerEntity.AlamedaScalerName != nil {
				alamedaScalerName = *containerEntity.AlamedaScalerName
			}
			if containerEntity.NodeName != nil {
				nodeName = *containerEntity.NodeName
			}
			if containerEntity.PodCreatedTime != nil {
				podCreatedTime = *containerEntity.PodCreatedTime
			}
			if containerEntity.Policy != nil {
				policy = *containerEntity.Policy
			}
			if containerEntity.Namespace != nil {
				topControllerNamespace = *containerEntity.Namespace
			}
			if containerEntity.TopControllerName != nil {
				topControllerName = *containerEntity.TopControllerName
			}
			if containerEntity.TopControllerKind != nil {
				topControllerKind = *containerEntity.TopControllerKind
			}
			if containerEntity.TpoControllerReplicas != nil {
				topControllerReplicas = *containerEntity.TpoControllerReplicas
			}
			if containerEntity.UsedRecommendationID != nil {
				usedRecommendationID = *containerEntity.UsedRecommendationID
			}

			if containerEntity.AppName != nil {
				appName = *containerEntity.AppName
			}
			if containerEntity.AppPartOf != nil {
				appPartOf = *containerEntity.AppPartOf
			}
			// TODO: handle compatibility of enableHPA and enableVPA
				if containerEntity.EnableHPA != nil {
					enableHPA = *containerEntity.EnableHPA
				}
				if containerEntity.EnableVPA != nil {
					enableVPA = *containerEntity.EnableVPA
				}
			if containerEntity.AlamedaScalerResourceLimitCPU != nil {
				alamedaScalerResourceLimitCPU = strconv.FormatFloat(*containerEntity.AlamedaScalerResourceLimitCPU, 'f', -1, 64)
			}
			if containerEntity.AlamedaScalerResourceLimitMemory != nil {
				alamedaScalerResourceLimitMemory = strconv.FormatFloat(*containerEntity.AlamedaScalerResourceLimitMemory, 'f', -1, 64)
			}
			if containerEntity.AlamedaScalerResourceRequestCPU != nil {
				alamedaScalerResourceRequestCPU = strconv.FormatFloat(*containerEntity.AlamedaScalerResourceRequestCPU, 'f', -1, 64)
			}
			if containerEntity.AlamedaScalerResourceRequestMemory != nil {
				alamedaScalerResourceRequestMemory = strconv.FormatFloat(*containerEntity.AlamedaScalerResourceRequestMemory, 'f', -1, 64)
			}

			datahubPod = &ApiResources.Pod{
				ObjectMeta: &ApiResources.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					NodeName:  nodeName,
					// TODO: cluster name
				},
				ResourceLink: resourceLink,
				Containers:   make([]*ApiResources.Container, 0),
				StartTime: &timestamp.Timestamp{
					Seconds: podCreatedTime,
				},
				TopController: &ApiResources.Controller{
					ObjectMeta: &ApiResources.ObjectMeta{
						Name:      topControllerName,
						Namespace: topControllerNamespace,
					},
					Replicas: topControllerReplicas,
					Kind:     FormatConvert.KindEnum[topControllerKind],
				},
				Status: &ApiResources.PodStatus{
					Phase:   ApiResources.PodPhase(ApiResources.PodPhase_value[podPhase]),
					Message: podMessage,
					Reason:  podReason,
				},
				AppName:   appName,
				AppPartOf: appPartOf,
				AlamedaPodSpec: &ApiResources.AlamedaPodSpec{
					AlamedaScaler: &ApiResources.ObjectMeta{
						Name:      alamedaScalerName,
						Namespace: alamedaScalerNamespace,
					},
					Policy:               FormatConvert.RecommendationPolicyEnum[policy],
					UsedRecommendationId: usedRecommendationID,
					AlamedaScalerResources: &ApiResources.ResourceRequirements{
						Limits: map[int32]string{
							int32(ApiCommon.ResourceName_CPU):    alamedaScalerResourceLimitCPU,
							int32(ApiCommon.ResourceName_MEMORY): alamedaScalerResourceLimitMemory,
						},
						Requests: map[int32]string{
							int32(ApiCommon.ResourceName_CPU):    alamedaScalerResourceRequestCPU,
							int32(ApiCommon.ResourceName_MEMORY): alamedaScalerResourceRequestMemory,
						},
					},
					// TODO: handle ScalingTool from Enable_VPA and Enable_HPA
					ScalingTool: ApiResources.ScalingTool_VPA,
				},
			}
			datahubPodMap[podID] = datahubPod
		}

		datahubContainer := containerEntityToDatahubContainer(containerEntity)
		datahubPod.Containers = append(datahubPod.Containers, datahubContainer)
	}

	for _, datahubPod := range datahubPodMap {
		copyDatahubPod := datahubPod
		datahubPods = append(datahubPods, copyDatahubPod)
	}

	return datahubPods
}
*/

/*
func containerEntityToDatahubContainer(containerEntity *EntityInfluxClusterStatus.ContainerEntity) *ApiResources.Container {

	var (
		statusWaitingReason                       string
		statusWaitingMessage                      string
		statusRunningStartedAt                    int64
		statusTerminatedExitCode                  int32
		statusTerminatedReason                    string
		statusTerminatedMessage                   string
		statusTerminatedStartedAt                 int64
		statusTerminatedFinishedAt                int64
		lastTerminationStatusWaitingReason        string
		lastTerminationStatusWaitingMessage       string
		lastTerminationStatusRunningStartedAt     int64
		lastTerminationStatusTerminatedExitCode   int32
		lastTerminationStatusTerminatedReason     string
		lastTerminationStatusTerminatedMessage    string
		lastTerminationStatusTerminatedStartedAt  int64
		lastTerminationStatusTerminatedFinishedAt int64
		restartCount                              int32
	)

	if containerEntity.StatusWaitingReason != nil {
		statusWaitingReason = *containerEntity.StatusWaitingReason
	}
	if containerEntity.StatusWaitingMessage != nil {
		statusWaitingMessage = *containerEntity.StatusWaitingMessage
	}
	if containerEntity.StatusRunningStartedAt != nil {
		statusRunningStartedAt = *containerEntity.StatusRunningStartedAt
	}
	if containerEntity.StatusTerminatedExitCode != nil {
		statusTerminatedExitCode = *containerEntity.StatusTerminatedExitCode
	}
	if containerEntity.StatusTerminatedReason != nil {
		statusTerminatedReason = *containerEntity.StatusTerminatedReason
	}
	if containerEntity.StatusTerminatedMessage != nil {
		statusTerminatedMessage = *containerEntity.StatusTerminatedMessage
	}
	if containerEntity.StatusTerminatedStartedAt != nil {
		statusTerminatedStartedAt = *containerEntity.StatusTerminatedStartedAt
	}
	if containerEntity.StatusTerminatedFinishedAt != nil {
		statusTerminatedFinishedAt = *containerEntity.StatusTerminatedFinishedAt
	}
	if containerEntity.LastTerminationStatusWaitingReason != nil {
		lastTerminationStatusWaitingReason = *containerEntity.LastTerminationStatusWaitingReason
	}
	if containerEntity.LastTerminationStatusWaitingMessage != nil {
		lastTerminationStatusWaitingMessage = *containerEntity.LastTerminationStatusWaitingMessage
	}
	if containerEntity.LastTerminationStatusRunningStartedAt != nil {
		lastTerminationStatusRunningStartedAt = *containerEntity.LastTerminationStatusRunningStartedAt
	}
	if containerEntity.LastTerminationStatusTerminatedExitCode != nil {
		lastTerminationStatusTerminatedExitCode = *containerEntity.LastTerminationStatusTerminatedExitCode
	}
	if containerEntity.LastTerminationStatusTerminatedReason != nil {
		lastTerminationStatusTerminatedReason = *containerEntity.LastTerminationStatusTerminatedReason
	}
	if containerEntity.LastTerminationStatusTerminatedMessage != nil {
		lastTerminationStatusTerminatedMessage = *containerEntity.LastTerminationStatusTerminatedMessage
	}
	if containerEntity.LastTerminationStatusTerminatedStartedAt != nil {
		lastTerminationStatusTerminatedStartedAt = *containerEntity.LastTerminationStatusTerminatedStartedAt
	}
	if containerEntity.LastTerminationStatusTerminatedFinishedAt != nil {
		lastTerminationStatusTerminatedFinishedAt = *containerEntity.LastTerminationStatusTerminatedFinishedAt
	}
	if containerEntity.RestartCount != nil {
		restartCount = *containerEntity.RestartCount
	}

	// Pack container
	container := &ApiResources.Container{}
	container.Name = *containerEntity.Name
	container.Resources = &ApiResources.ResourceRequirements{}
	container.Resources.Limits = make(map[int32]string)
	container.Resources.Requests = make(map[int32]string)
	if containerEntity.ResourceLimitCPU != nil {
		container.Resources.Limits[int32(ApiCommon.ResourceName_CPU)] = strconv.FormatFloat(*containerEntity.ResourceLimitCPU, 'f', -1, 64)
	}
	if containerEntity.ResourceLimitMemory != nil {
		container.Resources.Limits[int32(ApiCommon.ResourceName_MEMORY)] = strconv.FormatInt(*containerEntity.ResourceLimitMemory, 10)
	}
	if containerEntity.ResourceRequestCPU != nil {
		container.Resources.Requests[int32(ApiCommon.ResourceName_CPU)] = strconv.FormatFloat(*containerEntity.ResourceRequestCPU, 'f', -1, 64)
	}
	if containerEntity.ResourceRequestMemory != nil {
		container.Resources.Requests[int32(ApiCommon.ResourceName_MEMORY)] = strconv.FormatInt(*containerEntity.ResourceRequestMemory, 10)
	}
	containerStatus := &ApiResources.ContainerStatus{
		State: &ApiResources.ContainerState{
			Waiting: &ApiResources.ContainerStateWaiting{
				Reason:  statusWaitingReason,
				Message: statusWaitingMessage,
			},
			Running: &ApiResources.ContainerStateRunning{
				StartedAt: &timestamp.Timestamp{Seconds: statusRunningStartedAt},
			},
			Terminated: &ApiResources.ContainerStateTerminated{
				ExitCode:   statusTerminatedExitCode,
				Reason:     statusTerminatedReason,
				Message:    statusTerminatedMessage,
				StartedAt:  &timestamp.Timestamp{Seconds: statusTerminatedStartedAt},
				FinishedAt: &timestamp.Timestamp{Seconds: statusTerminatedFinishedAt},
			},
		},
		LastTerminationState: &ApiResources.ContainerState{
			Waiting: &ApiResources.ContainerStateWaiting{
				Reason:  lastTerminationStatusWaitingReason,
				Message: lastTerminationStatusWaitingMessage,
			},
			Running: &ApiResources.ContainerStateRunning{
				StartedAt: &timestamp.Timestamp{Seconds: lastTerminationStatusRunningStartedAt},
			},
			Terminated: &ApiResources.ContainerStateTerminated{
				ExitCode:   lastTerminationStatusTerminatedExitCode,
				Reason:     lastTerminationStatusTerminatedReason,
				Message:    lastTerminationStatusTerminatedMessage,
				StartedAt:  &timestamp.Timestamp{Seconds: lastTerminationStatusTerminatedStartedAt},
				FinishedAt: &timestamp.Timestamp{Seconds: lastTerminationStatusTerminatedFinishedAt},
			},
		},
		RestartCount: restartCount,
	}
	container.Status = containerStatus

	return container
}
*/

/*
func getDatahubPodIDString(containerEntity *EntityInfluxClusterStatus.ContainerEntity) string {

	var (
		namespace         string
		podName           string
		alamedaScalerNS   string
		alamedaScalerName string
		nodeName          string
	)

	if containerEntity.Namespace != nil {
		namespace = *containerEntity.Namespace
	}
	if containerEntity.PodName != nil {
		podName = *containerEntity.PodName
	}
	if containerEntity.AlamedaScalerNamespace != nil {
		alamedaScalerNS = *containerEntity.AlamedaScalerNamespace
	}
	if containerEntity.AlamedaScalerName != nil {
		alamedaScalerName = *containerEntity.AlamedaScalerName
	}
	if containerEntity.NodeName != nil {
		nodeName = *containerEntity.NodeName
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s", namespace, podName, alamedaScalerNS, alamedaScalerName, nodeName)
}
*/
