# prometheus operator

## install helm 3
```text
# install helm 3
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
# add & download repo
helm repo add azure	http://mirror.azure.cn/kubernetes/charts/
helm pull azure/prometheus-operator
# unarchived & install to specific namespace
kubectl create ns monitoring
kubectl apply -f crds/
helm install prometheus -n monitoring .
```
