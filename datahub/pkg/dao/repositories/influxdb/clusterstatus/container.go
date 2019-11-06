package clusterstatus

import (
	"fmt"
	EntityInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/entities/influxdb/clusterstatus"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	FormatConvert "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

var (
	scope = Log.RegisterScope("cluster_status_db_measurement", "cluster_status DB measurement", 0)
)

// ContainerRepository is used to operate node measurement of cluster_status database
type ContainerRepository struct {
	influxDB *InternalInflux.InfluxClient
}

// IsTag checks the column is tag or not
func (containerRepository *ContainerRepository) IsTag(column string) bool {
	for _, tag := range EntityInfluxClusterStatus.ContainerTags {
		if column == string(tag) {
			return true
		}
	}
	return false
}

// NewContainerRepository creates the ContainerRepository instance
func NewContainerRepository(influxDBCfg *InternalInflux.Config) *ContainerRepository {
	return &ContainerRepository{
		influxDB: &InternalInflux.InfluxClient{
			Address:  influxDBCfg.Address,
			Username: influxDBCfg.Username,
			Password: influxDBCfg.Password,
		},
	}
}

// ListAlamedaContainers list predicted containers have relation with arguments
func (containerRepository *ContainerRepository) ListAlamedaContainers(namespace, name string, kind ApiResources.Kind, timeRange *ApiCommon.TimeRange) ([]*ApiResources.Pod, error) {
	pods := []*ApiResources.Pod{}
	whereStr := ""

	conditions := make([]string, 0)
	relationStatement := ""
	podCreatePeriodCondition := containerRepository.getPodCreatePeriodCondition(timeRange)
	switch kind {
	case ApiResources.Kind_POD:
		if namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerNamespace, namespace))
		}
		if name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerPodName, name))
		}
		if podCreatePeriodCondition != "" {
			conditions = append(conditions, podCreatePeriodCondition)
		}
	case ApiResources.Kind_DEPLOYMENT:
		conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerKind, FormatConvert.KindDisp[ApiResources.Kind_DEPLOYMENT]))
		if namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerNamespace, namespace))
		}
		if name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerName, name))
		}
		if podCreatePeriodCondition != "" {
			conditions = append(conditions, podCreatePeriodCondition)
		}
	case ApiResources.Kind_DEPLOYMENTCONFIG:
		conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerKind, FormatConvert.KindDisp[ApiResources.Kind_DEPLOYMENTCONFIG]))
		if namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerNamespace, namespace))
		}
		if name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerName, name))
		}
		if podCreatePeriodCondition != "" {
			conditions = append(conditions, podCreatePeriodCondition)
		}
	case ApiResources.Kind_ALAMEDASCALER:
		if namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerAlamedaScalerNamespace, namespace))
		}
		if name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerAlamedaScalerName, name))
		}
		if podCreatePeriodCondition != "" {
			conditions = append(conditions, podCreatePeriodCondition)
		}
	case ApiResources.Kind_STATEFULSET:
		conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerKind, FormatConvert.KindDisp[ApiResources.Kind_STATEFULSET]))
		if namespace != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerNamespace, namespace))
		}
		if name != "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" = '%s'`, EntityInfluxClusterStatus.ContainerTopControllerName, name))
		}
		if podCreatePeriodCondition != "" {
			conditions = append(conditions, podCreatePeriodCondition)
		}
	default:
		return pods, errors.Errorf("no mapping filter statement with Datahub Kind: %s, skip building relation statement", ApiResources.Kind_name[int32(kind)])
	}
	if len(conditions) > 0 {
		relationStatement = fmt.Sprintf("(%s", conditions[0])
		for _, condition := range conditions[1:] {
			relationStatement += fmt.Sprintf(" AND %s", condition)
		}
		relationStatement += ")"
	}
	if relationStatement != "" {
		if whereStr != "" {
			whereStr = fmt.Sprintf("%s AND %s", whereStr, relationStatement)
		} else {
			whereStr = fmt.Sprintf("WHERE %s", relationStatement)
		}
	}

	cmd := fmt.Sprintf("SELECT * FROM %s %s GROUP BY \"%s\",\"%s\",\"%s\",\"%s\"",
		string(Container), whereStr,
		string(EntityInfluxClusterStatus.ContainerNamespace), string(EntityInfluxClusterStatus.ContainerPodName),
		string(EntityInfluxClusterStatus.ContainerAlamedaScalerNamespace), string(EntityInfluxClusterStatus.ContainerAlamedaScalerName))
	scope.Debugf("ListAlamedaContainers command: %s", cmd)
	if results, err := containerRepository.influxDB.QueryDB(cmd, string(RepoInflux.ClusterStatus)); err == nil {

		containerEntities := make([]*EntityInfluxClusterStatus.ContainerEntity, 0)

		rows := InternalInflux.PackMap(results)
		for _, row := range rows {
			for _, data := range row.Data {
				entity := EntityInfluxClusterStatus.NewContainerEntityFromMap(data)
				containerEntities = append(containerEntities, &entity)
			}
		}

		pods = buildDatahubPodsFromContainerEntities(containerEntities)
		return pods, nil
	} else {
		return pods, err
	}
}

// CreateContainers add containers information container measurement
func (containerRepository *ContainerRepository) CreateContainers(pods []*ApiResources.Pod) error {
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
}

// DeleteContainers set containers' field is_deleted to true into container measurement
func (containerRepository *ContainerRepository) DeleteContainers(pods []*ApiResources.Pod) error {
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
}

// ListPodsContainers list containers information container measurement
func (containerRepository *ContainerRepository) ListPodsContainers(pods []*ApiResources.Pod) ([]*EntityInfluxClusterStatus.ContainerEntity, error) {

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
}

func (containerRepository *ContainerRepository) getPodCreatePeriodCondition(timeRange *ApiCommon.TimeRange) string {
	if timeRange == nil {
		return ""
	}

	var start int64 = 0
	var end int64 = 0

	if timeRange.StartTime != nil {
		start = timeRange.StartTime.Seconds
	}

	if timeRange.EndTime != nil {
		end = timeRange.EndTime.Seconds
	}

	if start == 0 && end == 0 {
		return ""
	} else if start == 0 && end != 0 {
		period := fmt.Sprintf(`"pod_create_time" < %d`, end)
		return period
	} else if start != 0 && end == 0 {
		period := fmt.Sprintf(`"pod_create_time" >= %d`, start)
		return period
	} else if start != 0 && end != 0 {
		period := fmt.Sprintf(`"pod_create_time" >= %d AND "pod_create_time" < %d`, start, end)
		return period
	}

	return ""
}

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
			val = ApiResources.PodPhase_name[int32(ApiResources.PodPhase_Unknown)]
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
			/*
				if containerEntity.EnableHPA != nil {
					enableHPA = *containerEntity.EnableHPA
				}
				if containerEntity.EnableVPA != nil {
					enableVPA = *containerEntity.EnableVPA
				}
			*/
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
					ScalingTool: ApiResources.ScalingTool_SCALING_TOOL_VPA,
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
