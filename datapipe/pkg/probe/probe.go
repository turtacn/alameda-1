package probe

import (
	"github.com/containers-ai/alameda/pkg/utils/log"
	"os"
)

var scope = log.RegisterScope("probe", "datapipe health probe", 0)

func LivenessProbe(cfg *LivenessProbeConfig) {
	bindAddr := cfg.BindAddr
	err := pingDatapipe(bindAddr)
	if err != nil {
		scope.Errorf("Failed to do liveness probe due to %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func ReadinessProbe(cfg *ReadinessProbeConfig) {
	bindAddr := cfg.BindAddr
	err := queryApiServer(bindAddr)
	if err != nil {
		scope.Errorf("Failed to do readiness probe due to %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
