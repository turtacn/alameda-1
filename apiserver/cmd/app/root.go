package app

import (
	"errors"
	"strings"

	Config "github.com/containers-ai/alameda/apiserver/pkg/config"
	"github.com/containers-ai/alameda/cmd/app"

	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "apiserver",
	Short: "alameda apiserver",
	Long:  "",
}

var (
	configurationFilePath string

	scope  *log.Scope
	config Config.Config
)

func init() {
	RootCmd.AddCommand(RunCmd)
	RootCmd.AddCommand(app.VersionCmd)
	RootCmd.AddCommand(ProbeCmd)

	RootCmd.PersistentFlags().StringVar(&configurationFilePath, "config", "/etc/alameda/apiserver/apiserver.yml", "The path to apiserver configuration file.")
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

func mergeConfigFileValueWithDefaultConfigValue() {

	if configurationFilePath == "" {

	} else {

		viper.SetConfigFile(configurationFilePath)
		err := viper.ReadInConfig()
		if err != nil {
			panic(errors.New("Failed to read configuration file: " + err.Error()))
		}
		err = viper.Unmarshal(&config)
		if err != nil {
			panic(errors.New("Failed to unmarshal configuration file: " + err.Error()))
		}
	}
}

func initConfig() {

	config = Config.NewDefaultConfig()

	initViperSetting()
	mergeConfigFileValueWithDefaultConfigValue()
}

func initViperSetting() {
	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
}

func initLogger() {
	opt := log.DefaultOptions()
	opt.RotationMaxSize = defaultRotationMaxSizeMegabytes
	opt.RotationMaxBackups = defaultRotationMaxBackups
	opt.RotateOutputPath = defaultLogRotateOutputFile
	err := log.Configure(opt)
	if err != nil {
		panic(err)
	}

	scope = log.RegisterScope("apiserver_probe", "api server probe command", 0)
}
