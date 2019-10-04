package probe

import (
	"fmt"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	InternalRabbitMQ "github.com/containers-ai/alameda/internal/pkg/message-queue/rabbitmq"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"os/exec"
	"strings"
)

type ReadinessProbeConfig struct {
	InfluxdbCfg   *InternalInflux.Config
	PrometheusCfg *InternalPromth.Config
	RabbitMQCfg   *InternalRabbitMQ.Config
}

func queryInfluxdb(influxdbConfig *InternalInflux.Config) error {
	err := pingInfluxdb(influxdbConfig.Address)
	if err != nil {
		return errors.Wrap(err, "failed to ping to influxdb")
	}
	return nil
}

func queryPrometheus(prometheusConfig *InternalPromth.Config) error {
	emr := "exceeded maximum resolution"
	options := []DBCommon.Option{}

	podContainerCPURepo := RepoPromthMetric.NewPodContainerCpuUsagePercentageRepositoryWithConfig(*prometheusConfig)
	containerCPUEntities, err := podContainerCPURepo.ListMetricsByPodNamespacedName("", "", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list pod metrics failed")
	}

	if err == nil && len(containerCPUEntities) == 0 {
		return fmt.Errorf("no container CPU metric found")
	}

	podContainerMemoryRepo := RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(*prometheusConfig)
	containerMemoryEntities, err := podContainerMemoryRepo.ListMetricsByPodNamespacedName("", "", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list pod metrics failed")
	}

	if err == nil && len(containerMemoryEntities) == 0 {
		return fmt.Errorf("no container memory metric found")
	}

	nodeCPUUsageRepo := RepoPromthMetric.NewNodeCpuUsagePercentageRepositoryWithConfig(*prometheusConfig)
	nodeCPUUsageEntities, err := nodeCPUUsageRepo.ListMetricsByNodeName("", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list node cpu usage metrics failed")
	}

	if err == nil && len(nodeCPUUsageEntities) == 0 {
		return fmt.Errorf("no node CPU metric found")
	}

	nodeMemoryUsageRepo := RepoPromthMetric.NewNodeMemoryBytesUsageRepositoryWithConfig(*prometheusConfig)
	nodeMemoryUsageEntities, err := nodeMemoryUsageRepo.ListMetricsByNodeName("", options...)
	if err != nil && !strings.Contains(err.Error(), emr) {
		return errors.Wrap(err, "list node memory usage metrics failed")
	}

	if err == nil && len(nodeMemoryUsageEntities) == 0 {
		return fmt.Errorf("no node memory metric found")
	}

	return nil
}

func queryQueue(rabbitmqConfig *InternalRabbitMQ.Config) error {
	return connQueue(rabbitmqConfig.URL)
}

func connQueue(url string) error {
	_, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	return nil
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
