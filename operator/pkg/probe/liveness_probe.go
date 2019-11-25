package probe

import (
	"os/exec"
)

type LivenessProbeConfig struct {
	ValidationSvc *ValidationSvc
}

type ValidationSvc struct {
	SvcName string
	SvcNS   string
	SvcPort int32
}

func queryWebhookSvc(svcURL string) error {
	curlCmd := exec.Command("curl", "-k", svcURL)

	_, err := curlCmd.CombinedOutput()
	if err != nil {
		return err
	}

	return err
}
