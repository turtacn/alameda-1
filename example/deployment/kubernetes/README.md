# Deploy Alameda on a Kubernetes Cluster

## Prerequisites  
- **A running Kubernetest cluster**  
It is recommended to use Kubernetes 1.9 or greater. Please follow [Kubernetes setup](https://kubernetes.io/docs/setup/) document to setup a solution that fits your needs. If you just want to have a taste of Alameda, you can start from [kubeadm](https://kubernetes.io/docs/setup/independent/install-kubeadm/) to setup a test cluster on a Ubuntu machine in 6 steps.
    - Install [kubeadm](https://kubernetes.io/docs/setup/independent/install-kubeadm/#k8s-install-0)
    - Disable swap in order for the kubelet to work properly by
    ```
    swapoff -a
    ```
    - Initialize master node by 
    ```
    kubeadm init --pod-network-cidr=10.244.0.0/16
    ```
    - Make *kubectl* work for your non-root user by
    ```
    mkdir -p $HOME/.kube
    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown $(id -u):$(id -g) $HOME/.kube/config
    ```
    - Allow pod scheduling on the master
    ```
    kubectl taint nodes --all node-role.kubernetes.io/master-
    ```
    - Install Flannel pod network add-on to allow pods communicate with each other by
    ```
    sysctl net.bridge.bridge-nf-call-iptables=1
    kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/bc79dd1505b0c8681ece4de4c0d86c5cd2643275/Documentation/kube-flannel.yml
    ```
    
- **A docker registry where to download Alameda component images**  
For now, Alameda component images are not yet in any public docker registry such as [docker hub](https://hub.docker.com/). We assume they are already built in local docker repositories and pulled by Kubernetes locally (All nodes in the cluster should have the same pre-pulled images). Please refer to [Build](https://github.com/containers-ai/alameda/blob/master/docs/build.md) document for how to build Alameda component images and [pre-pulled images](https://kubernetes.io/docs/concepts/containers/images/#pre-pulling-images) for how Kubernetes pulls images from local docker repositories.

As showed in the [Alameda architecture design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), Alameda includes several components. The following gives an example to deploy them.

## Deploy Prometheus

Alameda leverages [Prometheus](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-usage-monitoring/#prometheus) to collect Kubernetes metrics. For those clusters without a running Prometheus, please deploy one with commands:
```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f 3rdParty/prometheus.yaml
```

> **Note**: This example prometheus deployment use HTTP communication without any authentication.

For those clusters that already have a running Prometheus, please make sure that:
- cAdvisor metrics is scraped, and
- [Node exporter](https://github.com/prometheus/node_exporter) metrics is scraped, and
- [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) metrics is scraped, and
- recording rules required by Alameda is installed.

And the following block are recording rules required by Alameda.
```
groups:
- name: k8s.rules
  rules:
  - expr: |
      sum(rate(container_cpu_usage_seconds_total{image!="", container_name!=""}[5m])) by (namespace)
    record: namespace:container_cpu_usage_seconds_total:sum_rate
  - expr: |
      sum by (namespace, pod_name, container_name) (
        rate(container_cpu_usage_seconds_total{container_name!=""}[5m])
      )
    record: namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate
  - expr: |
      sum(container_memory_usage_bytes{image!="", container_name!=""}) by (namespace)
    record: namespace:container_memory_usage_bytes:sum
    - expr: max by (kubernetes_node, kubernetes_namespace, kubernetes_pod) 
        (label_replace(
          label_replace(
            label_replace(kube_pod_info{job="kubernetes-service-endpoints"}, "kubernetes_node", "$1", "node", "(.*)"),
          "kubernetes_namespace", "$1", "namespace", "(.*)"),
        "kubernetes_pod", "$1", "pod", "(.*)"))
      record: "node_namespace_pod:kube_pod_info:"
    - expr: label_replace(node_cpu_seconds_total, "cpu", "$1", "cpu", "cpu(.+)")
      record: node_cpu
    - expr: label_replace(1 - avg by (kubernetes_node) (rate(node_cpu{job="kubernetes-service-endpoints",mode="idle"}[1m]) * on(kubernetes_namespace, kubernetes_pod) group_left(node) node_namespace_pod:kube_pod_info:), "node", "$1", "kubernetes_node", "(.*)")
      record: node:node_cpu_utilisation:avg1m
    - expr: node_memory_MemTotal_bytes
      record: node_memory_MemTotal
    - expr: node_memory_MemFree_bytes
      record: node_memory_MemFree
    - expr: node_memory_Cached_bytes
      record: node_memory_Cached
    - expr: node_memory_Buffers_bytes
      record: node_memory_Buffers
    - expr: label_replace(sum
        by (kubernetes_node) ((node_memory_MemFree{job="kubernetes-service-endpoints"} + node_memory_Cached{job="kubernetes-service-endpoints"}
        + node_memory_Buffers{job="kubernetes-service-endpoints"}) * on(kubernetes_namespace, kubernetes_pod) group_left(kubernetes_node)
        node_namespace_pod:kube_pod_info:), "node", "$1", "kubernetes_node", "(.*)")
      record: node:node_memory_bytes_available:sum
    - expr: label_replace(sum
        by(kubernetes_node) (node_memory_MemTotal{job="kubernetes-service-endpoints"} * on(kubernetes_namespace, kubernetes_pod)
        group_left(kubernetes_node) node_namespace_pod:kube_pod_info:), "node", "$1", "kubernetes_node", "(.*)")
      record: node:node_memory_bytes_total:sum
```

## Deploy Operator, Datahub and Alameda-ai Components

In the next, we will:
- Create a *alameda* namespace, service accounts and required RBAC settings
- Install custom resource definitions (CRDs) to the Kubernetes cluster
- Deploy the operator component to interact with Kubernetes components
- Deploy the datahub component to access Prometheus and InfluxDB
- Deploy the alameda-ai component for computations of metrics predictions and orchestration recommendations

Please execute the following commands.
```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f crds/
$ kubectl create -f deployconfig/
```
> **Note**: If you are using your own Prometheus, please make sure the container environment variables of deployconfig/datahub.yaml are properly set to connect it.

## Deploy InfluxDB component

This component stores the predicted metrics and recommendations. Please follow the following commands to install it.

```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f 3rdParty/influxdb.yaml
``` 

> **Note**: If you are using your own InfluxDB, please make sure the container environment variables of deployconfig/datahub.yaml are properly set to connect it.

## Deploy Grafana component (optional)

This component visualizes the predicted metrics. This is optional and can be ignored if you do not need it.
Alameda core function can work without this Grafana component. Please follow the following commands as your needs.

```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f 3rdParty/grafana.yaml
``` 

