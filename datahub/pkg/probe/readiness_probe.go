package probe

import (
	"fmt"
	"github.com/streadway/amqp"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

type ReadinessProbeConfig struct {
	InfluxdbAddr  string
	PrometheusCfg *InternalPromth.Config
	RabbitMQCfg  *InternalRabbitMQ.Config
}

func pingInfluxdb(influxdbAddr string) error {
	pingURL := fmt.Sprintf("%s/ping", influxdbAddr)
	curlCmd := exec.Command("curl", "-sl", "-I", pingURL)
	if strings.Contains(pingURL, "https") {
		curlCmd = exec.Command("curl", "-sl", "-I", "-k", pingURL)
	}
	_, err := curlCmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func queryPrometheus(prometheusConfig *InternalPromth.Config) error {
	emr := "exceeded maximum resolution"
	options := []DBCommon.Option{}

	podContainerCPURepo := RepoPromthMetric.NewPodContainerCPUUsagePercentageRepositoryWithConfig(*prometheusConfig)
	containerCPUEntities, err := podContainerCPURepo.ListMetricsByPodNamespacedName("", "", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list pod metrics failed")
	}

	if err == nil && len(containerCPUEntities) == 0 {
		return fmt.Errorf("No container CPU metric found")
	}

	podContainerMemoryRepo := RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(*prometheusConfig)
	containerMemoryEntities, err := podContainerMemoryRepo.ListMetricsByPodNamespacedName("", "", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list pod metrics failed")
	}

	if err == nil && len(containerMemoryEntities) == 0 {
		return fmt.Errorf("No container memory metric found")
	}

	nodeCPUUsageRepo := RepoPromthMetric.NewNodeCPUUsagePercentageRepositoryWithConfig(*prometheusConfig)
	nodeCPUUsageEntities, err := nodeCPUUsageRepo.ListMetricsByNodeName("", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list node cpu usage metrics failed")
	}

	if err == nil && len(nodeCPUUsageEntities) == 0 {
		return fmt.Errorf("No node CPU metric found")
	}

	nodeMemoryUsageRepo := RepoPromthMetric.NewNodeMemoryUsageBytesRepositoryWithConfig(*prometheusConfig)
	nodeMemoryUsageEntities, err := nodeMemoryUsageRepo.ListMetricsByNodeName("", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list node memory usage metrics failed")
	}

	if err == nil && len(nodeMemoryUsageEntities) == 0 {
		return fmt.Errorf("No node memory metric found")
	}

	return nil
}

func connQueue(url string) error {
	_, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	return nil
}
