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
	datahubAddr := cfg.DatahubAddr
	err := queryDatahub(datahubAddr)
	if err != nil {
		scope.Errorf("Readiness probe: query datahub failed due to %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
