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
	"io/ioutil"
	"os"
	"strings"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_serializer_json "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"

	"github.com/containers-ai/alameda/operator"
	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	"github.com/containers-ai/alameda/operator/pkg/apis"
	"github.com/containers-ai/alameda/operator/pkg/controller"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/operator/pkg/webhook"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	appsapi "github.com/openshift/api/apps"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
var crdLocation string

var operatorConf operator.Config
var scope *logUtil.Scope

var (
	// VERSION is sofeware version
	VERSION string
	// BUILD_TIME is build time
	BUILD_TIME string
	// GO_VERSION is go version
	GO_VERSION string
)

func init() {
	flag.BoolVar(&isLogOutput, "logfile", false, "output log file")
	flag.IntVar(&serverPort, "server-port", 50050, "Local gRPC server port")
	flag.StringVar(&operatorConfigFile, "config", "/etc/alameda/operator/operator.yml", "File path to operator coniguration")
	flag.StringVar(&crdLocation, "crd-location", "/etc/alameda/operator/crds", "CRD location")

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
	applyCRDs(cfg)
	initServerConfig(mgr)
	initLogger()
	printSoftwareInfo()

	scope.Info("Registering Components.")
	registerThirdPartyCRD()
	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		scope.Error(err.Error())
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		scope.Error(err.Error())
	}

	scope.Info("setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		scope.Errorf("unable to register webhooks to the manager: %s", err.Error())
		os.Exit(1)
	}

	go registerNodes(mgr.GetClient())
	scope.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		scope.Error(err.Error())
	}
}

func registerNodes(client client.Client) {
	time.Sleep(3 * time.Second)
	listResources := resources.NewListResources(client)
	nodes, err := listResources.ListAllNodes()
	if err != nil {
		scope.Errorf("register nodes to Datahub failed: %s", err.Error())
		return
	}
	scope.Infof(fmt.Sprintf("%v nodes found in cluster.", len(nodes)))
	datahubNodeRepo := datahub_node.NewAlamedaNodeRepository()
	datahubNodeRepo.CreateAlamedaNode(nodes)
}

func registerThirdPartyCRD() {
	apis.AddToSchemes = append(apis.AddToSchemes, appsapi.Install)
}

func printSoftwareInfo() {
	scope.Infof(fmt.Sprintf("Alameda Version: %s", VERSION))
	scope.Infof(fmt.Sprintf("Alameda Build Time: %s", BUILD_TIME))
	scope.Infof(fmt.Sprintf("Alameda GO Version: %s", GO_VERSION))
}

func applyCRDs(cfg *rest.Config) {
	apiextensionsClientSet, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	crdFiles := []string{}
	if files, err := ioutil.ReadDir(crdLocation); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				crdFiles = append(crdFiles, crdLocation+string(os.PathSeparator)+file.Name())
			}
		}
	} else {
		scope.Error("Failed to read CRDs: " + err.Error())
	}

	for _, crdFile := range crdFiles {
		yamlBin, rfErr := ioutil.ReadFile(crdFile)
		if rfErr != nil {
			scope.Errorf(fmt.Sprintf("Read crd file %s failed.", crdFile))
			continue
		}

		s := k8s_serializer_json.NewYAMLSerializer(k8s_serializer_json.DefaultMetaFactory, scheme.Scheme,
			scheme.Scheme)

		var crdIns apiextensionsv1beta1.CustomResourceDefinition
		_, _, decErr := s.Decode(yamlBin, nil, &crdIns)
		if decErr != nil {
			scope.Errorf(fmt.Sprintf("Decode crd file %s failed: %s", crdFile, decErr.Error()))
			continue
		}

		_, createErr := apiextensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crdIns)
		if createErr != nil {
			scope.Errorf(fmt.Sprintf("Failed to create CRD %s: %s", crdFile, createErr.Error()))
			continue
		}
		err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
			crd, getErr := apiextensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdIns.Name, metav1.GetOptions{})
			if getErr != nil {
				scope.Warnf(fmt.Sprintf("Failed to wait for CRD %s creation: %s", crdFile, getErr.Error()))
				return false, nil
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextensionsv1beta1.Established:
					if cond.Status == apiextensionsv1beta1.ConditionTrue {
						scope.Infof(fmt.Sprintf("CRD %s created.", crdIns.Name))
						return true, nil
					}
				case apiextensionsv1beta1.NamesAccepted:
					if cond.Status == apiextensionsv1beta1.ConditionFalse {
						scope.Errorf(fmt.Sprintf("CRD name conflict: %v, %v", cond.Reason, err))
					}
				}
			}
			return false, nil
		})
		if err != nil {
			scope.Errorf(fmt.Sprintf("Polling crd for $s failed: %s", crdFile, err.Error()))
		}
	}
}
