package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/containers-ai/alameda/cmd/app"
	"github.com/containers-ai/alameda/datahub"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envVarPrefix = "ALAMEDA_DATAHUB"
)

var (
	scope  *log.Scope
	config datahub.Config

	configurationFilePath string

	RunCmd = &cobra.Command{
		Use:   "run",
		Short: "start alameda datahub server",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {

			var (
				err error

				server *datahub.Server
			)
			app.PrintSoftwareVer()
			initConfig()
			initLogger()
			setLoggerScopesWithConfig(*config.Log)
			displayConfig()
			server, err = datahub.NewServer(config)
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

func init() {
	parseFlag()
}

func parseFlag() {
	RunCmd.Flags().StringVar(&configurationFilePath, "config", "/etc/alameda/datahub/datahub.yml", "The path to datahub configuration file.")
}

func initConfig() {

	config = datahub.NewDefaultConfig()

	initViperSetting()
	mergeConfigFileValueWithDefaultConfigValue()
}

func initViperSetting() {

	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
}

func mergeConfigFileValueWithDefaultConfigValue() {

	if configurationFilePath == "" {

	} else {

		viper.SetConfigFile(configurationFilePath)
		err := viper.ReadInConfig()
		if err != nil {
			panic(errors.New("Read configuration file failed: " + err.Error()))
		}
		err = viper.Unmarshal(&config)
		if err != nil {
			panic(errors.New("Unmarshal configuration failed: " + err.Error()))
		}
	}
}

func initLogger() {

	scope = log.RegisterScope("datahub", "datahub server log", 0)
}

func displayConfig() {
	if configBin, err := json.MarshalIndent(config, "", "  "); err != nil {
		scope.Error(err.Error())
	} else {
		scope.Infof(fmt.Sprintf("Datahub configuration: %s", string(configBin)))
	}
}

func setLoggerScopesWithConfig(config log.Config) {
	for _, scope := range log.Scopes() {
		scope.SetLogCallers(config.SetLogCallers == true)
		if outputLvl, ok := log.StringToLevel(config.OutputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := log.StringToLevel(config.StackTraceLevel); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}
