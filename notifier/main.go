/*

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
	"os"
	"strings"

	notifyingv1alpha1 "github.com/containers-ai/alameda/notifier/api/v1alpha1"
	"github.com/containers-ai/alameda/notifier/controllers"
	"github.com/containers-ai/alameda/notifier/probe"
	"github.com/containers-ai/alameda/notifier/queue"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	// +kubebuilder:scaffold:imports
)

var (
	scheme              = runtime.NewScheme()
	logRotateOutputFile string
	scope               *log.Scope
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = notifyingv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

var (
	configFile         string
	readinessProbeFlag bool
	livenessProbeFlag  bool
)

const (
	envVarPrefix                    = "ALAMEDA_NOTIFIER"
	defaultRotationMaxSizeMegabytes = 100
	defaultRotationMaxBackups       = 7
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&configFile, "config", "/etc/alameda/notifier/notifier.yml", "File path to notifier coniguration")
	flag.StringVar(&logRotateOutputFile, "log-output-file", "/var/log/alameda/alameda-ai-notifier.log", "The path of log file.")
	flag.BoolVar(&readinessProbeFlag, "readiness-probe", false, "probe for readiness")
	flag.BoolVar(&livenessProbeFlag, "liveness-probe", false, "probe for liveness")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	initConfig()
	initLogger()
	setLoggerScopesWithConfig()
	if readinessProbeFlag && livenessProbeFlag {
		scope.Error("Cannot run readiness probe and liveness probe at the same time")
		return
	} else if readinessProbeFlag {
		probe.ReadinessProbe(&probe.ReadinessProbeConfig{})
		return
	} else if livenessProbeFlag {
		probe.LivenessProbe(&probe.LivenessProbeConfig{})
		return
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})

	if err != nil {
		scope.Errorf("unable to start manager: %s", err.Error())
		os.Exit(1)
	}

	if err = (&controllers.AlamedaNotificationTopicReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf("unable to create controller: %s", err.Error())
		os.Exit(1)
	}
	if err = (&controllers.AlamedaNotificationChannelReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf("unable to create controller: %s", err.Error())
		os.Exit(1)
	}

	if err = (&notifyingv1alpha1.AlamedaNotificationChannel{}).SetupWebhookWithManager(mgr); err != nil {
		scope.Errorf("unable to create webhook: %s", err.Error())
		os.Exit(1)
	}
	if err = (&notifyingv1alpha1.AlamedaNotificationTopic{}).SetupWebhookWithManager(mgr); err != nil {
		scope.Errorf("unable to create webhook: %s", err.Error())
		os.Exit(1)
	}
	if viper.IsSet("k8sWebhook.port") {
		whSrv := mgr.GetWebhookServer()
		whSrv.Port = viper.GetInt("k8sWebhook.port")
	}

	// +kubebuilder:scaffold:builder
	go launchQueueConsumer(mgr)

	scope.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		scope.Errorf("problem running manager: %s", err.Error())
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initLogger() {

	opt := log.DefaultOptions()
	opt.RotationMaxSize = defaultRotationMaxSizeMegabytes
	opt.RotationMaxBackups = defaultRotationMaxBackups
	opt.RotateOutputPath = logRotateOutputFile

	err := log.Configure(opt)
	if err != nil {
		panic(err)
	}

	scope = log.RegisterScope("app", "notifier app", 0)
}

func setLoggerScopesWithConfig() {
	setLogCallers := true
	if viper.IsSet("log.setLogcallers") {
		setLogCallers = viper.GetBool("log.setLogcallers")
	}

	outputLevel := "none"
	if viper.IsSet("log.outputLevel") {
		outputLevel = viper.GetString("log.outputLevel")
	}

	stackTraceLvl := "none"
	if viper.IsSet("log.stacktraceLevel") {
		stackTraceLvl = viper.GetString("log.stacktraceLevel")
	}
	for _, scope := range log.Scopes() {
		scope.SetLogCallers(setLogCallers)
		if outputLvl, ok := log.StringToLevel(outputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := log.StringToLevel(stackTraceLvl); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}

func launchQueueConsumer(mgr manager.Manager) {
	queueURL := viper.GetString("rabbitmq.url")
	qc := queue.NewRabbitMQClient(mgr, queueURL)
	qc.Start()
}
