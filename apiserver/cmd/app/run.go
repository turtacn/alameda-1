package app

import (
	"encoding/json"
	"fmt"
	"github.com/containers-ai/alameda/apiserver"
	"github.com/containers-ai/alameda/cmd/app"
	"github.com/spf13/cobra"
)

const (
	envVarPrefix                    = "ALAMEDA_API"
	defaultRotationMaxSizeMegabytes = 100
	defaultRotationMaxBackups       = 7
	defaultLogRotateOutputFile      = "/var/log/alameda/alameda-api.log"
)

var (
	RunCmd = &cobra.Command{
		Use:   "run",
		Short: "start alameda api server",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {

			var (
				err error

				server *apiserver.Server
			)
			app.PrintSoftwareVer()
			initConfig()
			initLogger()
			setLoggerScopesWithConfig(*config.Log)
			displayConfig()
			server, err = apiserver.NewServer(config)
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
		scope.Infof(fmt.Sprintf("Alameda API server configuration: %s", string(configBin)))
	}
}
