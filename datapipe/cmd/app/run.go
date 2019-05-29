package app

import (
	"encoding/json"
	"fmt"
	"github.com/containers-ai/alameda/cmd/app"
	"github.com/containers-ai/alameda/datapipe"
	"github.com/spf13/cobra"
)

const (
	envVarPrefix                    = "ALAMEDA_DATAPIPE"
	defaultRotationMaxSizeMegabytes = 100
	defaultRotationMaxBackups       = 7
	defaultLogRotateOutputFile      = "/var/log/alameda/alameda-datapipe.log"
)

var (
	RunCmd = &cobra.Command{
		Use:   "run",
		Short: "start alameda datapipe",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {

			var (
				err error

				server *datapipe.Server
			)
			app.PrintSoftwareVer()
			initConfig()
			initLogger()
			setLoggerScopesWithConfig(*config.Log)
			displayConfig()
			server, err = datapipe.NewServer(config)
			if err != nil {
				panic(err)
			}

			if err = server.Run(); err != nil {
				server.Stop()
				panic(err)
			}
		},
	}
)

func displayConfig() {
	if configBin, err := json.MarshalIndent(config, "", "  "); err != nil {
		scope.Error(err.Error())
	} else {
		scope.Infof(fmt.Sprintf("Datapipe configuration: %s", string(configBin)))
	}
}
