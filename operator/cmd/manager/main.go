/*
Copyright 2019 The Alameda Authors.

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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/containers-ai/alameda/operator"
	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	"github.com/containers-ai/alameda/operator/pkg/apis"
	"github.com/containers-ai/alameda/operator/pkg/controller"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/operator/server"
	grafanadatasource "github.com/containers-ai/alameda/operator/services/grafana-datasource"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	envVarPrefix = "ALAMEDA"
)

const JSONIndent = "  "

var isLogOutput bool
var serverPort int
var operatorConfigFile string

var operatorConf operator.Config
var scope *logUtil.Scope
var wg sync.WaitGroup

func init() {
	flag.BoolVar(&isLogOutput, "logfile", false, "output log file")
	flag.IntVar(&serverPort, "server-port", 50050, "Local gRPC server port")
	flag.StringVar(&operatorConfigFile, "config", "/etc/alameda/operator/operator.yml", "File path to operator coniguration")

	scope = logUtil.RegisterScope("manager", "operator entry point", 0)
}

func initLogger() {
	scope.Infof("Log output level is %s.", operatorConf.Log.OutputLevel)
	scope.Infof("Log stacktrace level is %s.", operatorConf.Log.StackTraceLevel)
	for _, scope := range logUtil.Scopes() {
		scope.SetLogCallers(operatorConf.Log.SetLogCallers == true)
		if outputLvl, ok := logUtil.StringToLevel(operatorConf.Log.OutputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := logUtil.StringToLevel(operatorConf.Log.StackTraceLevel); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}

func initServerConfig(mgr manager.Manager) {

	operatorConf = operator.NewConfig(mgr)

	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// TODO: This config need default value. And it should check the file exists befor SetConfigFile.
	viper.SetConfigFile(operatorConfigFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(errors.New("Read configuration failed: " + err.Error()))
	}
	err = viper.Unmarshal(&operatorConf)
	if err != nil {
		panic(errors.New("Unmarshal configuration failed: " + err.Error()))
	} else {
		if operatorConfBin, err := json.MarshalIndent(operatorConf, "", JSONIndent); err == nil {
			scope.Infof(fmt.Sprintf("Operator configuration: %s", string(operatorConfBin)))
		}
	}
}

func main() {
	flag.Parse()

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
	// TODO: There are config dependency, this manager should have it's config.
	initServerConfig(mgr)
	initLogger()

	// Setup grpc server config
	s, err := server.NewServer(&operatorConf)

	if err != nil {
		scope.Error("Setup server failed: " + err.Error())
	}

	// Start server
	wg.Add(1)
	go s.Start(&wg)
	go watchServer(s)

	scope.Info("Registering Components.")
	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		scope.Error(err.Error())
	}
	go grafanadatasource.NewGrafanaDataSource(mgr, operatorConf.GrafanaBackend.BindPort)
	go registerNodes(mgr.GetClient())
	scope.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		scope.Error(err.Error())
	}

	// Wait grpc server goroutine
	wg.Wait()
}

func watchServer(s *server.Server) {
	var err error

	select {
	case err = <-s.Err():
		s.Close(&wg)
	}
	scope.Error("server runtime failed: " + err.Error())
}

func registerNodes(client client.Client) {
	time.Sleep(3 * time.Second)
	listResources := resources.NewListResources(client)
	nodeList, _ := listResources.ListAllNodes()
	scope.Infof(fmt.Sprintf("%v nodes found in cluster.", len(nodeList)))
	createAlamedaNode := datahub_node.NewCreateAlamedaNode()
	createAlamedaNode.CreateAlamedaNode(nodeList)
}
