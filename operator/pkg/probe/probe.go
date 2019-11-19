package probe

import (
	"fmt"
	"os"

	"github.com/containers-ai/alameda/pkg/utils/log"
)

var scope = log.RegisterScope("probe", "datahub health probe", 0)

func LivenessProbe(cfg *LivenessProbeConfig) {
	svcName := cfg.ValidationSvc.SvcName
	svcNS := cfg.ValidationSvc.SvcNS
	svcPort := cfg.ValidationSvc.SvcPort
	svcURL := fmt.Sprintf("https://%s.%s:%s", svcName, svcNS, fmt.Sprint(svcPort))
	err := queryWebhookSvc(svcURL)
	if err != nil {
		scope.Errorf("Liveness probe: query validation webhook service %s failed due to %s",
			svcURL, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func ReadinessProbe(cfg *ReadinessProbeConfig) {
	datahubAddr := cfg.DatahubAddr
	err := queryDatahub(datahubAddr)
	if err != nil {
		scope.Errorf("Readiness probe: query datahub %s failed due to %s",
			datahubAddr, err.Error())
		os.Exit(1)
	}

	svcURL := fmt.Sprintf("https://localhost:%s", fmt.Sprint(cfg.WHSrvPort))
	err = queryWebhookSrv(svcURL)
	if err != nil {
		scope.Errorf("Readiness probe: query validation webhook server %s failed due to %s",
			svcURL, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
