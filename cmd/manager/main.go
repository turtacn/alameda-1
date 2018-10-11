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
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"

	"github.com/containers-ai/alameda/pkg/apis"
	as_clientset "github.com/containers-ai/alameda/pkg/client/clientset/versioned"
	as_lister "github.com/containers-ai/alameda/pkg/client/listers/autoscaling/v1alpha1"
	"github.com/containers-ai/alameda/pkg/controller"
	utils_alameda "github.com/containers-ai/alameda/pkg/utils/alameda"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	inCluster  bool
	kubeconfig *string
)

func init() {
	flag.BoolVar(&inCluster, "in-cluster", true, "Kubernetes client in cluster")
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func main() {
	kube_flag.InitFlags()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		glog.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		glog.Fatal(err)
	}

	glog.V(2).Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		glog.Fatal(err)
	}

	glog.V(2).Info("Starting the Cmd.")

	stopChannel := make(chan struct{})
	startVPAController(stopChannel)

	// Start the Cmd
	glog.Fatal(mgr.Start(signals.SetupSignalHandler()))
}
func startVPAController(stopChannel <-chan struct{}) {
	vpaLister := newReadyASLister(stopChannel)
	vpascalers, _ := vpaLister.List(labels.Everything())
	glog.V(2).Info(len(vpascalers), " vpa scalers.")
}

func newReadyASLister(stopChannel <-chan struct{}) as_lister.AlamedaVPALister {
	var cfg *restclient.Config
	var err error

	if inCluster {
		cfg, err = inClusterConfig()
		glog.V(2).Info("Kubernetes client is in cluster.")
	} else {
		cfg, err = outClusterConfig()
		glog.V(2).Info("Kubernetes client is out cluster.")
	}

	if err != nil {
		panic(err.Error())
	}

	autoscalingClient, err := as_clientset.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	return utils_alameda.NewAllVpasLister(autoscalingClient, stopChannel)
}

func inClusterConfig() (*restclient.Config, error) {
	// creates the in-cluster config
	config, err := restclient.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	return config, err
}

func outClusterConfig() (*restclient.Config, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config, err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
