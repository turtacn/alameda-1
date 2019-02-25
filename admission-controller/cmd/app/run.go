package app

import (
	"errors"
	"flag"
	"net/http"
	"strings"

	"github.com/containers-ai/alameda/admission-controller"
	"github.com/containers-ai/alameda/admission-controller/pkg/recommendator/resource/datahub"
	"github.com/containers-ai/alameda/admission-controller/pkg/server"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envVarPrefix  = "ALAMEDA_ADMCTL"
	allowEmptyEnv = true
)

var (
	scope  *log.Scope
	config *admission_controller.Config

	configurationFilePath string

	RunCmd = &cobra.Command{
		Use:   "run",
		Short: "start alameda admission-controller server",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {

			flag.Parse()

			initConfig()
			initLog()
			datahubResourceRecommendator, err := datahub.NewDatahubResourceRecommendatorWithConfig(config.Datahub)
			if err != nil {
				panic(err.Error())
			}
			admissionController, err := server.NewAdmissionControllerWithConfig(server.Config{Enable: config.Enable}, datahubResourceRecommendator)
			if err != nil {
				panic(err.Error())
			}

			mux := http.NewServeMux()
			registerHandlerFunc(mux, admissionController)

			server := newHTTPServer(*config, mux)
			server.ListenAndServeTLS("", "")
		},
	}
)

func init() {
	flag.StringVar(&configurationFilePath, "config", "/etc/alameda/admission-controller/admission-controller.yml", "File path to admission-controller coniguration")
}

func initConfig() {

	defaultConfig := admission_controller.NewDefaultConfig()
	config = &defaultConfig
	initViperSetting()
	mergeConfigFileValueWithDefaultConfigValue()
}

func initViperSetting() {

	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// viper.AllowEmptyEnv(allowEmptyEnv)
}

func mergeConfigFileValueWithDefaultConfigValue() {

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

func initLog() {
	for _, scope := range log.Scopes() {
		scope.SetLogCallers(config.Log.SetLogCallers == true)
		if outputLvl, ok := log.StringToLevel(config.Log.OutputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := log.StringToLevel(config.Log.StackTraceLevel); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}

func registerHandlerFunc(mux *http.ServeMux, ac server.AdmissionController) {
	mux.HandleFunc("/pods", ac.MutatePod)
}

func newHTTPServer(cfg admission_controller.Config, mux *http.ServeMux) *http.Server {

	clientset := admission_controller.GetK8SClient()

	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: admission_controller.ConfigTLS(cfg, clientset),
	}

	return server
}
