package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/containers-ai/alameda/cmd/app"
	"github.com/containers-ai/alameda/evictioner"
	"github.com/containers-ai/alameda/evictioner/pkg/eviction"
	"github.com/containers-ai/alameda/operator/pkg/apis"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s_config "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	envVarPrefix = "ALAMEDA_EVICTIONER"
)

var (
	scope  *log.Scope
	config evictioner.Config

	configurationFilePath string

	RunCmd = &cobra.Command{
		Use:   "run",
		Short: "start alameda evictioner",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			app.PrintSoftwareVer()
			initConfig()
			initLogger()
			setLoggerScopesWithConfig(*config.Log)
			displayConfig()
			startEvictioner()
		},
	}
)

func init() {
	parseFlag()
}

func parseFlag() {
	RunCmd.Flags().StringVar(&configurationFilePath, "config", "/etc/alameda/evictioner/evictioner.yml", "The path to evictioner configuration file.")
}

func initConfig() {

	config = evictioner.NewDefaultConfig()

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

	scope = log.RegisterScope("evict", "evict server log", 0)
}

func displayConfig() {
	if configBin, err := json.MarshalIndent(config, "", "  "); err != nil {
		scope.Error(err.Error())
	} else {
		scope.Infof(fmt.Sprintf("Evict configuration: %s", string(configBin)))
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

func startEvictioner() {
	conn, err := grpc.Dial(config.Datahub.Address, grpc.WithInsecure())
	if err != nil {
		scope.Errorf("create pods to datahub failed: %s", err.Error())
		return
	}

	defer conn.Close()

	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

	k8sClientConfig, err := k8s_config.GetConfig()
	if err != nil {
		scope.Error("Get kubernetes configuration failed: " + err.Error())
		return
	}

	k8sCli, err := client.New(k8sClientConfig, client.Options{})
	if err != nil {
		scope.Error("Create kubernetes client failed: " + err.Error())
		return
	}

	mgr, err := manager.New(k8sClientConfig, manager.Options{})
	if err != nil {
		scope.Error(err.Error())
	}
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	evictioner := eviction.NewEvictioner(config.Eviction.CheckCycle,
		datahubServiceClnt,
		k8sCli,
		*config.Eviction,
	)
	evictioner.Start()
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
