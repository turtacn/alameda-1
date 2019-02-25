package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/containers-ai/alameda/admission-controller/pkg/recommendator/resource/datahub"
	"github.com/containers-ai/alameda/admission-controller/pkg/server"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	envVarPrefix  = "ALAMEDA_ADMCTL"
	allowEmptyEnv = true
)

var (
	configurationFilePath string
	config                *Config
)

func init() {
	flag.StringVar(&configurationFilePath, "config", "/etc/alameda/admission-controller/admission-controller.yml", "File path to admission-controller coniguration")
}

func initConfig() {

	defaultConfig := NewDefaultConfig()
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

func main() {

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
}

func registerHandlerFunc(mux *http.ServeMux, ac server.AdmissionController) {
	mux.HandleFunc("/pods", ac.MutatePod)
}

func newHTTPServer(cfg Config, mux *http.ServeMux) *http.Server {

	clientset := getClient()

	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: configTLS(cfg, clientset),
	}

	return server
}
