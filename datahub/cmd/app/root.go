package app

import (
	"errors"
	"github.com/containers-ai/alameda/cmd/app"
	Keycodes "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	DatahubConfig "github.com/containers-ai/alameda/datahub/pkg/config"
	Notifier "github.com/containers-ai/alameda/datahub/pkg/notifier"
	EventMgt "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

var RootCmd = &cobra.Command{
	Use:   "datahub",
	Short: "alameda datahub",
	Long:  "",
}

var (
	configurationFilePath string

	scope  *log.Scope
	config DatahubConfig.Config
)

func init() {
	RootCmd.AddCommand(RunCmd)
	RootCmd.AddCommand(app.VersionCmd)
	RootCmd.AddCommand(ProbeCmd)

	RootCmd.PersistentFlags().StringVar(&configurationFilePath, "config", "/etc/alameda/datahub/datahub.toml", "The path to datahub configuration file.")
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
	config = DatahubConfig.NewDefaultConfig()
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

	scope = log.RegisterScope("datahub_probe", "datahub probe command", 0)
}

func initEventMgt() {
	scope.Info("Initialize event management")

	EventMgt.InitEventMgt(config.InfluxDB, config.RabbitMQ)
}

func initKeycode() {
	scope.Info("Initialize keycode management")

	Keycodes.KeycodeInit(config.Keycode)
	keycodeMgt := Keycodes.NewKeycodeMgt()
	keycodeMgt.Refresh(true)
}

func initNotifier() {
	scope.Info("Initialize notifier")

	Notifier.NotifierInit(config.Notifier)
	go Notifier.Run()
}
