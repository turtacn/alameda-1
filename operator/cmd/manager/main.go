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
	"github.com/spf13/viper"
	"log"
	"strings"
	"sync"

	"github.com/containers-ai/alameda/operator/pkg/apis"
	"github.com/containers-ai/alameda/operator/pkg/controller"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"github.com/containers-ai/alameda/operator/server"
	"github.com/go-logr/logr"
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

var logger logr.Logger
var serverConf server.Config
var scope *logUtil.Scope

func init() {
	flag.BoolVar(&isDev, "development", false, "development mode")
	flag.BoolVar(&isLogOutput, "logfile", false, "output log file")
	flag.StringVar(&aiSrvAddr, "ai-server", "alameda-ai.alameda.svc.cluster.local:50051", "AI service address")
	flag.IntVar(&serverPort, "server-port", 50050, "Local gRPC server port")
	flag.StringVar(&operatorConfigFile, "config", "/etc/alameda/operator/operator.yml", "File path to operator coniguration")

	scope = logUtil.RegisterScope("manager", "operator entry point", 0)
}

func initLogger(development bool) {
	logger = logUtil.GetLogger()
}

func initConfig(mgr manager.Manager) {

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

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if !flag.Parsed() {
		flag.Parse()
	}
	initLogger(isDev)

	if err != nil {
		scope.Error("Get configuration failed: " + err.Error())
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		logUtil.GetLogger().Error(err, "Create manager failed.")
	}

	// Set wait group for Server goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	initConfig(mgr)

	// Setup Server
	s, err := server.NewServer(&serverConf)
	if err != nil {
		scope.Error("Setup server failed: " + err.Error())
	}

	// Start Server
	go s.Start(&wg)
	log.Printf("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logUtil.GetLogger().Error(err, "Add scheme failed.")
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logUtil.GetLogger().Error(err, "Add controller failed.")
	}

	log.Printf("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logUtil.GetLogger().Error(err, "Run manager failed.")
	}

	// Wait Server goroutine
	wg.Wait()
}
