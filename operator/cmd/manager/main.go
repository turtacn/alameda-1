/*
Copyright 2018 The Alameda Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"strings"
	"sync"

	"github.com/containers-ai/alameda/operator/pkg/apis"
	"github.com/containers-ai/alameda/operator/pkg/controller"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"github.com/containers-ai/alameda/operator/server"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	envVarPrefix = "ALAMEDA_OPERATOR"
)

var isDev bool
var aiSrvAddr string
var isLogOutput bool
var serverPort int
var operatorConfigFile string

var serverConf server.Config
var scope *logUtil.Scope
var wg sync.WaitGroup

func init() {
	flag.BoolVar(&isDev, "development", false, "development mode")
	flag.BoolVar(&isLogOutput, "logfile", false, "output log file")
	flag.StringVar(&aiSrvAddr, "ai-server", "alameda-ai.alameda.svc.cluster.local:50051", "AI service address")
	flag.IntVar(&serverPort, "server-port", 50050, "Local gRPC server port")
	flag.StringVar(&operatorConfigFile, "config", "/etc/alameda/operator/operator.yml", "File path to operator coniguration")

	scope = logUtil.RegisterScope("manager", "operator entry point", 0)
}

func initLogger() {
	scope.Infof("Log output level is %s.", serverConf.Log.OutputLevel)
	scope.Infof("Log stacktrace level is %s.", serverConf.Log.StackTraceLevel)
	for _, scope := range logUtil.Scopes() {
		scope.SetLogCallers(serverConf.Log.SetLogCallers == true)
		if outputLvl, ok := logUtil.StringToLevel(serverConf.Log.OutputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := logUtil.StringToLevel(serverConf.Log.StackTraceLevel); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}

func initServerConfig(mgr manager.Manager) {

	serverConf = server.NewConfig(mgr)

	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.SetConfigFile(operatorConfigFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(errors.New("Read configuration failed: " + err.Error()))
	}
	viper.Unmarshal(&serverConf)
}

func main() {
	flag.Parse()

	initLogger()
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		scope.Error("Get configuration failed: " + err.Error())
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		scope.Error(err.Error())
	}

	initServerConfig(mgr)

	// Setup grpc server config
	s, err := server.NewServer(&serverConf)

	if err != nil {
		scope.Error("Setup server failed: " + err.Error())
	}

	// Start grpc server
	wg.Add(1)
	go s.Start(&wg)

	scope.Info("Registering Components.")
	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		scope.Error(err.Error())
	}

	scope.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		scope.Error(err.Error())
	}

	// Wait grpc server goroutine
	wg.Wait()
}
