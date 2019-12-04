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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	"github.com/containers-ai/alameda/operator/controllers"
	datahub_client_application "github.com/containers-ai/alameda/operator/datahub/client/application"
	datahub_client_controller "github.com/containers-ai/alameda/operator/datahub/client/controller"
	datahub_client_namespace "github.com/containers-ai/alameda/operator/datahub/client/namespace"
	datahub_client_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	datahub_client_pod "github.com/containers-ai/alameda/operator/datahub/client/pod"
	"github.com/containers-ai/alameda/operator/pkg/probe"
	"github.com/containers-ai/alameda/operator/pkg/utils"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources/validate"
	op_webhook "github.com/containers-ai/alameda/operator/pkg/webhook"
	"github.com/containers-ai/alameda/pkg/provider"
	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahubv1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	osappsapi "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	envVarPrefix = "ALAMEDA_OPERATOR"

	defaultRotationMaxSizeMegabytes = 100
	defaultRotationMaxBackups       = 7
	defaultLogRotateOutputFile      = "/var/log/alameda/alameda-operator.log"
)

const JSONIndent = "  "

var operatorConfigFile string
var crdLocation string
var showVer bool
var readinessProbeFlag bool
var livenessProbeFlag bool
var metricsAddr string
var enableLeaderElection bool

var operatorConf operator.Config
var k8sConfig *rest.Config
var scope *logUtil.Scope
var clusterUID string

var (
	datahubConn   *grpc.ClientConn
	datahubClient datahubv1alpha1.DatahubServiceClient
)

var (
	// VERSION is sofeware version
	VERSION string
	// BUILD_TIME is build time
	BUILD_TIME string
	// GO_VERSION is go version
	GO_VERSION string

	scheme = runtime.NewScheme()
)

func init() {
	flag.BoolVar(&showVer, "version", false, "show version")
	flag.BoolVar(&readinessProbeFlag, "readiness-probe", false, "probe for readiness")
	flag.BoolVar(&livenessProbeFlag, "liveness-probe", false, "probe for liveness")
	flag.StringVar(&operatorConfigFile, "config", "/etc/alameda/operator/operator.toml",
		"File path to operator coniguration")
	flag.StringVar(&crdLocation, "crd-location", "/etc/alameda/operator/crds", "CRD location")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	scope = logUtil.RegisterScope("manager", "operator entry point", 0)

	_ = clientgoscheme.AddToScheme(scheme)

	_ = autoscalingv1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	if ok, _ := utils.ServerHasOpenshiftAPIAppsV1(); ok {
		_ = osappsapi.AddToScheme(scheme)
	}
	// +kubebuilder:scaffold:scheme
}

func initLogger() {

	opt := logUtil.DefaultOptions()
	opt.RotationMaxSize = defaultRotationMaxSizeMegabytes
	logFilePath := viper.GetString("log.filePath")
	if logFilePath == "" {
		logFilePath = defaultLogRotateOutputFile
	}
	opt.RotationMaxBackups = defaultRotationMaxBackups
	opt.RotateOutputPath = logFilePath
	err := logUtil.Configure(opt)
	if err != nil {
		panic(err)
	}

	scope.Infof("Log output level is %s.", operatorConf.Log.OutputLevel)
	scope.Infof("Log stacktrace level is %s.", operatorConf.Log.StackTraceLevel)
	for _, scope := range logUtil.Scopes() {
		scope.SetLogCallers(operatorConf.Log.SetLogCallers == true)
		if outputLvl, ok := logUtil.StringToLevel(operatorConf.Log.OutputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok :=
			logUtil.StringToLevel(operatorConf.Log.StackTraceLevel); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}

func initServerConfig(mgr *manager.Manager) {

	operatorConf = operator.NewConfigWithoutMgr()
	if mgr != nil {
		operatorConf = operator.NewConfig(*mgr)
	}

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
		if operatorConfBin, err :=
			json.MarshalIndent(operatorConf, "", JSONIndent); err == nil {
			scope.Infof(fmt.Sprintf("Operator configuration: %s",
				string(operatorConfBin)))
		}
	}
}

func initThirdPartyClient() {
	for {
		datahubConn, _ = grpc.Dial(operatorConf.Datahub.Address,
			grpc.WithInsecure(), grpc.WithUnaryInterceptor(
				grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(uint(3)))))
		datahubClient = datahubv1alpha1.NewDatahubServiceClient(datahubConn)
		_, err := datahubClient.ListNodes(context.Background(), &datahub_resources.ListNodesRequest{})
		if err == nil {
			break
		} else {
			scope.Errorf("connect datahub failed on init: %s", err.Error())
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func initClusterUID() error {
	k8sClient, err := client.New(k8sConfig, client.Options{})
	if err != nil {
		return errors.Wrap(err, "new kubernetes client failed")
	}

	clusterUID, err = k8sutils.GetClusterUID(k8sClient)
	if err != nil {
		return errors.Wrap(err, "get cluster uid failed")
	} else if clusterUID == "" {
		return errors.New("get empty cluster uid")
	}

	return nil
}

func setupWebhook(mgr manager.Manager) error {

	scope.Info("Setting up webhooks")
	vr := &validate.ResourceValidate{}
	if err := (&autoscalingv1alpha1.AlamedaScaler{
		Validate: vr,
	}).SetupWebhookWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	whSrv := mgr.GetWebhookServer()
	deploymentValidatingHook := &webhook.Admission{
		Handler: admission.HandlerFunc(func(ctx context.Context,
			req webhook.AdmissionRequest) webhook.AdmissionResponse {
			decoder, err := admission.NewDecoder(scheme)
			if err != nil {
				scope.Errorf("new decoder failed %s", err.Error())
				return webhook.Denied(err.Error())
			}
			return op_webhook.HandleDeployment(decoder, mgr.GetClient(), ctx, req)
		}),
	}

	if viper.IsSet("k8sWebhookServer.admissionPaths.validateDeployment") {
		whSrv.Register(
			viper.GetString("k8sWebhookServer.admissionPaths.validateDeployment"),
			deploymentValidatingHook)
	}

	if ok, _ := utils.ServerHasOpenshiftAPIAppsV1(); ok {
		if viper.IsSet("k8sWebhookServer.admissionPaths.validateDeploymentConfig") {
			deploymentConfigValidatingHook := &webhook.Admission{
				Handler: admission.HandlerFunc(func(ctx context.Context,
					req webhook.AdmissionRequest) webhook.AdmissionResponse {
					decoder, err := admission.NewDecoder(scheme)
					if err != nil {
						scope.Errorf("new decoder failed %s", err.Error())
						return webhook.Denied(err.Error())
					}
					return op_webhook.HandleDeploymentConfig(decoder, mgr.GetClient(), ctx, req)
				}),
			}
			whSrv.Register(
				viper.GetString("k8sWebhookServer.admissionPaths.validateDeploymentConfig"),
				deploymentConfigValidatingHook)
		}
	}
	if viper.IsSet("k8sWebhookServer.port") {
		whSrv.Port = viper.GetInt("k8sWebhookServer.port")
	}

	return nil
}

func main() {
	flag.Parse()
	if showVer {
		printSoftwareInfo()
		return
	}

	if readinessProbeFlag && livenessProbeFlag {
		scope.Error("Cannot run readiness probe and liveness probe at the same time")
		return
	} else if readinessProbeFlag {
		initServerConfig(nil)
		opWHSrvPort := viper.GetInt32("k8sWebhookServer.port")
		readinessProbe(&probe.ReadinessProbeConfig{
			WHSrvPort:   opWHSrvPort,
			DatahubAddr: operatorConf.Datahub.Address,
		})
		return
	} else if livenessProbeFlag {
		initServerConfig(nil)
		opWHSrvName := viper.GetString("k8sWebhookServer.service.name")
		opWHSrvNamespace := viper.GetString("k8sWebhookServer.service.namespace")
		opWHSrvPort := viper.GetInt32("k8sWebhookServer.service.port")
		livenessProbe(&probe.LivenessProbeConfig{
			ValidationSvc: &probe.ValidationSvc{
				SvcName: opWHSrvName,
				SvcNS:   opWHSrvNamespace,
				SvcPort: opWHSrvPort,
			},
		})
		return
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		scope.Error("Get configuration failed: " + err.Error())
	}
	k8sConfig = cfg

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})

	if err != nil {
		scope.Error(err.Error())
		os.Exit(1)
	}

	// TODO: There are config dependency, this manager should have it's config.
	initServerConfig(&mgr)
	initLogger()
	printSoftwareInfo()
	initThirdPartyClient()
	err = initClusterUID()
	if err != nil {
		panic(err)
	}

	scope.Info("Registering Components.")
	datahubControllerRepo := datahub_client_controller.NewControllerRepository(datahubConn, clusterUID)
	datahubPodRepo := datahub_client_pod.NewPodRepository(datahubConn, clusterUID)
	datahubNamespaceRepo := datahub_client_namespace.NewNamespaceRepository(datahubConn, clusterUID)

	// ------------------------ Setup Controllers ------------------------
	if err = (&controllers.AlamedaScalerReconciler{
		Client:                 mgr.GetClient(),
		Scheme:                 mgr.GetScheme(),
		ClusterUID:             clusterUID,
		DatahubApplicationRepo: datahub_client_application.NewApplicationRepository(datahubConn, clusterUID),
		DatahubControllerRepo:  datahubControllerRepo,
		DatahubNamespaceRepo: datahubNamespaceRepo,
		DatahubPodRepo:         datahubPodRepo,
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	if err = (&controllers.AlamedaRecommendationReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ClusterUID:    clusterUID,
		DatahubClient: datahubv1alpha1.NewDatahubServiceClient(datahubConn),
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	if err = (&controllers.DeploymentReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		ClusterUID:            clusterUID,
		DatahubControllerRepo: datahubControllerRepo,
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	if ok, _ := utils.ServerHasOpenshiftAPIAppsV1(); ok {
		if err = (&controllers.DeploymentConfigReconciler{
			Client:                mgr.GetClient(),
			Scheme:                mgr.GetScheme(),
			ClusterUID:            clusterUID,
			DatahubControllerRepo: datahubControllerRepo,
		}).SetupWithManager(mgr); err != nil {
			scope.Errorf(err.Error())
			os.Exit(1)
		}
	}

	if err = (&controllers.NamespaceReconciler{
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		ClusterUID:           clusterUID,
		DatahubNamespaceRepo: datahubNamespaceRepo,
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	cloudprovider := ""
	if provider.OnGCE() {
		cloudprovider = provider.GCP
	} else if provider.OnEC2() {
		cloudprovider = provider.AWS
	}
	regionName := ""
	switch cloudprovider {
	case provider.AWS:
		regionName = provider.AWSRegionMap[provider.GetEC2Region()]
	}
	if err = (&controllers.NodeReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		ClusterUID:      clusterUID,
		Cloudprovider:   cloudprovider,
		RegionName:      regionName,
		DatahubNodeRepo: *datahub_client_node.NewNodeRepository(datahubConn, clusterUID),
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}

	if err = (&controllers.StatefulSetReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		ClusterUID:            clusterUID,
		DatahubControllerRepo: datahubControllerRepo,
	}).SetupWithManager(mgr); err != nil {
		scope.Errorf(err.Error())
		os.Exit(1)
	}
	// ------------------------ Setup Controllers ------------------------

	setupWebhook(mgr)

	wg, ctx := errgroup.WithContext(context.Background())
	wg.Go(
		func() error {
			// To use instance from return value of function mgr.GetClient(),
			// block till the cache is synchronized, or the cache will be empty and get/list nothing.
			ok := mgr.GetCache().WaitForCacheSync(ctx.Done())
			if !ok {
				scope.Error("Wait for cache synchronization failed")
			} else {
				go syncResourcesWithDatahub(mgr.GetClient(),
					datahubConn)
			}
			return nil
		})

	wg.Go(
		func() error {
			scope.Info("Starting the Cmd.")
			return mgr.Start(ctrl.SetupSignalHandler())
		})

	if err := wg.Wait(); err != nil {
		scope.Error(err.Error())
	}
}
