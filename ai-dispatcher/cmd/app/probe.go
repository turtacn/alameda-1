package app

import (
	"os"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/probe"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	PROBE_TYPE_READINESS = "readiness"
	PROBE_TYPE_LIVENESS  = "liveness"
)

var (
	probeType string

	ProbeCmd = &cobra.Command{
		Use:   "probe",
		Short: "probe alameda ai dispatcher",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			initLogger()
			setLoggerScopesWithConfig()
			startProbing()
		},
	}
)

func init() {
	parseProbeFlag()
}

func parseProbeFlag() {
	ProbeCmd.Flags().StringVar(&probeType, "type", PROBE_TYPE_READINESS, "The probe type for ai dispatcher.")
}

func startProbing() {
	datahubAddr := viper.GetString("datahubAddress")
	if probeType == PROBE_TYPE_LIVENESS {
		probe.LivenessProbe(&probe.LivenessProbeConfig{})
	} else if probeType == PROBE_TYPE_READINESS {
		probe.ReadinessProbe(&probe.ReadinessProbeConfig{
			DatahubAddr: datahubAddr,
			QueueURL:    viper.GetString("queue.url"),
		})
	} else {
		scope.Errorf("Probe type does not supports %s, please try %s or %s.", probeType, PROBE_TYPE_LIVENESS, PROBE_TYPE_READINESS)
		os.Exit(1)
	}
}
