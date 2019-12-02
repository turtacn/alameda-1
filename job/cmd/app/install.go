package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers-ai/alameda/job/pkg/assets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

var (
	InstallCmd = &cobra.Command{
		Use:   "install",
		Short: "install prometheus rules",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			initK8SClient()
			initConfig()
			for {
				if err := install(); err != nil {
					scope.Errorf("install rules failed: %s", err.Error())
					time.Sleep(time.Duration(1) * time.Second)
				} else {
					break
				}
			}
		},
	}
	retrySec = 1
)

func install() error {
	promDeployNS := "nks-system"
	promDeployName := "prometheus"
	promConfigCMName := "prometheus-config"
	promConfigCMKey := "prometheus.yaml"
	if viper.IsSet("retrySec") {
		retrySec = viper.GetInt("retrySec")
	}
	if viper.IsSet("prometheus.namespace") {
		promDeployNS = viper.GetString("prometheus.namespace")
	}
	if viper.IsSet("prometheus.name") {
		promDeployName = viper.GetString("prometheus.name")
	}
	if viper.IsSet("prometheus.configCMName") {
		promConfigCMName = viper.GetString("prometheus.configCMName")
	}
	if viper.IsSet("prometheus.cmConfigKey") {
		promConfigCMKey = viper.GetString("prometheus.cmConfigKey")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	promConfigMap := &corev1.ConfigMap{}
	if err := k8sCli.Get(ctx, k8sTypes.NamespacedName{
		Namespace: promDeployNS,
		Name:      promConfigCMName,
	}, promConfigMap); err != nil {
		return err
	}

	if strings.Contains(promConfigMap.Data[promConfigCMKey], "rule_files:") {
		scope.Infof("skip applying rule files because config already include rule files definitions")
		return nil
	}

	ruleFilesStr := "rule_files:\n"
	for _, ruleFileName := range assets.AssetNames() {
		ruleFileBin, err := assets.Asset(ruleFileName)
		ruleFileBaseName := filepath.Base(ruleFileName)
		ruleFilesStr = fmt.Sprintf("%s  - %s\n", ruleFilesStr, ruleFileBaseName)
		if err != nil {
			return err
		}
		promConfigMap.Data[ruleFileBaseName] = string(ruleFileBin)
	}
	ruleFilesStr = fmt.Sprintf("%s\n", ruleFilesStr)
	if !strings.Contains(promConfigMap.Data[promConfigCMKey], "rule_files:") {
		promConfigMap.Data[promConfigCMKey] = fmt.Sprintf("%s\n%s", promConfigMap.Data[promConfigCMKey], ruleFilesStr)
	}

	scope.Infof(promConfigMap.Data[promConfigCMKey])

	if err := k8sCli.Update(ctx, promConfigMap); err != nil {
		return err
	}

	for {
		if deploy, err := getPromDeploy(promDeployNS, promDeployName); err == nil {
			if *deploy.Spec.Replicas == int32(1) {
				scope.Infof("start scalng down prometheus deployment")
				rep := int32(0)
				deploy.Spec.Replicas = &rep
				if err := k8sCli.Update(ctx, deploy); err == nil {
					scope.Infof("scale download prometheus deployment successfully")
					break
				} else {
					scope.Errorf("failed to scale down prometheus: %s", err.Error())
				}
			}
		} else {
			scope.Errorf("failed to get prometheus to scale down: %s", err.Error())
		}
		time.Sleep(time.Duration(retrySec) * time.Second)
	}

	for {
		if deploy, err := getPromDeploy(promDeployNS, promDeployName); err == nil {
			if *deploy.Spec.Replicas == int32(0) {
				scope.Infof("start scaling up prometheus deployment")
				rep := int32(1)
				deploy.Spec.Replicas = &rep
				if err := k8sCli.Update(ctx, deploy); err == nil {
					scope.Infof("scale up prometheus deployment successfully")
					break
				} else {
					scope.Errorf("failed to scale up prometheus: %s", err.Error())
				}
			}
		} else {
			scope.Errorf("failed to get prometheus to scale up: %s", err.Error())
		}
		time.Sleep(time.Duration(retrySec) * time.Second)
	}

	return nil
}

func getPromDeploy(ns, name string) (*appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if viper.IsSet("retrySec") {
		retrySec = viper.GetInt("retrySec")
	}
	deploy := &appsv1.Deployment{}

	if err := k8sCli.Get(ctx, k8sTypes.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}
