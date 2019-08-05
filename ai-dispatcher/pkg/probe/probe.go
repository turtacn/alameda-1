package probe

import (
	"os"

	"github.com/containers-ai/alameda/pkg/utils/log"
)

var scope = log.RegisterScope("probe", "ai dispatcher health probe", 0)

func LivenessProbe(cfg *LivenessProbeConfig) {
	os.Exit(0)
}

func ReadinessProbe(cfg *ReadinessProbeConfig) {
	// query datahub
	datahubAddr := cfg.DatahubAddr
	err := queryDatahub(datahubAddr)
	if err != nil {
		scope.Errorf("Readiness probe: query datahub with address (%s) failed due to %s", datahubAddr, err.Error())
		os.Exit(1)
	}
	// connect queue
	err = connQueue(cfg.QueueURL)
	if err != nil {
		scope.Errorf("Readiness probe: query queue with url (%s) failed due to %s", cfg.QueueURL, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
