package app

import (
	"os"

	"github.com/containers-ai/alameda/datahub/pkg/probe"
	"github.com/spf13/cobra"
)

const (
	ProbeTypeLiveness  = "liveness"
	ProbeTypeReadiness = "readiness"
)

var (
	probeType string

	ProbeCmd = &cobra.Command{
		Use:   "probe",
		Short: "probe alameda datahub server",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			initLogger()
			setLoggerScopesWithConfig(*config.Log)
			startProbing()
		},
	}
)

func init() {
	parseProbeFlag()
}

func parseProbeFlag() {
	ProbeCmd.Flags().StringVar(&probeType, "type", ProbeTypeLiveness, "The probe type for datahub.")
}

func startProbing() {
	if probeType == ProbeTypeLiveness {
		probe.LivenessProbe(&probe.LivenessProbeConfig{
			BindAddr: config.BindAddress,
		})
	} else if probeType == ProbeTypeReadiness {
		probe.ReadinessProbe(&probe.ReadinessProbeConfig{
			InfluxdbAddr:  config.InfluxDB.Address,
			PrometheusCfg: config.Prometheus,
			RabbitMQCfg: config.RabbitMQ,
		})
	} else {
		scope.Errorf("Probe type does not supports %s, please try %s or %s.", probeType, ProbeTypeLiveness, ProbeTypeReadiness)
		os.Exit(1)
	}
}
