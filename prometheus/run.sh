#!/bin/sh
# install helm 3
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh

## 选择一，已经过期了
helm repo add azure http://mirror.azure.cn/kubernetes/charts/
helm pull azure/prometheus-operator
tar zxvf prometheus-operator-9.3.2.tgz
cd prometheus-operator
kubectl create ns monitoring
kubectl apply -f crds/
helm install prometheus -n monitoring .

## 其他安装选择，社区官方版
#helm repo add prometheus-community	https://prometheus-community.github.io/helm-charts/
#helm pull   prometheus-community/kube-prometheus-stack
#tar zxvf kube-prometheus-stack-14.3.0.tgz
#cd kube-prometheus-stack
#kubectl create ns monitoring
#kubectl apply -f crds/
#helm install prometheus -n monitoring .



