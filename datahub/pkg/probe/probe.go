package probe

import (
	"os"

	"github.com/containers-ai/alameda/pkg/utils/log"
)

var scope = log.RegisterScope("probe", "datahub health probe", 0)

func LivenessProbe(cfg *LivenessProbeConfig) {
	bindAddr := cfg.BindAddr
	err := queryDatahub(bindAddr)
	if err != nil {
		scope.Errorf("Liveness probe failed with address (%s) due to %s", bindAddr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func ReadinessProbe(cfg *ReadinessProbeConfig) {
	influxdbAddr := cfg.InfluxdbAddr
	prometheusCfg := cfg.PrometheusCfg
	queueCfg := cfg.RabbitMQCfg

	err := pingInfluxdb(influxdbAddr)
	if err != nil {
		scope.Errorf("Readiness probe: ping influxdb with address (%s) failed due to %s", influxdbAddr, err.Error())
		os.Exit(1)
	}

	err = queryPrometheus(prometheusCfg)
	if err != nil {
		scope.Errorf("Readiness probe: query prometheus failed with url (%s) due to %s", prometheusCfg.URL, err.Error())
		os.Exit(1)
	}

	err = connQueue(queueCfg.URL)
	if err != nil {
		scope.Errorf("Readiness probe: query queue failed with url (%s) due to %s", queueCfg.URL, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
