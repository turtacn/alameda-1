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
	"flag"
	"log"
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

var isDev bool
var logger logr.Logger
var serverConf server.Config

func init() {
	flag.BoolVar(&isDev, "development", false, "development mode")
}

func initLogger(development bool) {
	logger = logUtil.GetLogger()
}

func initConfig(mgr manager.Manager) {
	serverConf = server.NewConfig(mgr)
}

func main() {

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if !flag.Parsed() {
		flag.Parse()
	}
	initLogger(isDev)

	if err != nil {
		log.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Fatal(err)
	}
	// Set wait group for Server goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	initConfig(mgr)
	// Setup Server
	s, err := server.NewServer(&serverConf)
	if err != nil {
		log.Fatal(err)
	}

	// Start Server
	go s.Start(&wg)
	log.Printf("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")

	// Start the Cmd
	log.Fatal(mgr.Start(signals.SetupSignalHandler()))

	// Wait Server goroutine
	wg.Wait()
}
