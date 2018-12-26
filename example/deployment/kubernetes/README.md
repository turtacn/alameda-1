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
For now, Alameda component images are not yet in any public docker registry such as [docker hub](https://hub.docker.com/). We assume they are already built in local docker repositories and pulled by Kubernetes locally (All nodes in the cluster should have the same pre-pulled images). Please refer to [Build](https://github.com/containers-ai/alameda/blob/master/design/build.md) document for how to build Alameda component images and [pre-pulled images](https://kubernetes.io/docs/concepts/containers/images/#pre-pulling-images) for how Kubernetes pulls images from local docker repositories.

As showed in the [Alameda architecture design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), Alameda includes several components. The following gives an example to deploy them.

## Deploy Prometheus

Alameda leverages [Prometheus](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-usage-monitoring/#prometheus) to collect Kubernetes metrics. For those clusters without a running Prometheus, please deploy one with commands:
```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f 3rdParty/prometheus.yaml
```

> **Note**: This example prometheus deployment use HTTP communication without any authentication.

For those clusters that already have a running Prometheus, please make sure that:
- cAdvisor metrics is scraped and 
- recording rules required by Alameda is installed.

The following block gives an example configuration to scrape cAdvisor metrics of every node.
```
scrape_configs:
- job_name: kubernetes-cAdvisor
  scheme: https
  scrape_interval: 30s
  scrape_timeout: 10s
  metrics_path: /metrics
  kubernetes_sd_configs:
  - role: node
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token # for permission to query kubelet api
  tls_config: # for validate certificate from kubelet
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  relabel_configs:
  - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
  - target_label: __address__
    replacement: kubernetes.default.svc:443
  - source_labels: [__meta_kubernetes_node_name]
    regex: (.+)
    target_label: __metrics_path__
    replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor
```

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
```

## Deploy Operator and Alameda-ai Components

In the next, we will:
- Create a *alameda* namespace, service accounts and required RBAC settings associated to Alameda operator
- Install custom resource definitions (CRDs) to the Kubernetes cluster
- Deploy the operator component to interact with Kubernetes components
- Deploy the alameda-ai component for computations of metrics predictions and orchestration recommendations

Please execute the following commands.
```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f crds/
$ kubectl create -f deployconfig/
```
> **Note**: If you are using your own Prometheus, please make sure the container environment variables **ALAMEDA_GRPC_PROMETHEUS_URL** and **ALAMEDA_GRPC_PROMETHEUS_BEARER_TOKEN_FILE** in deployconfig/operator.yaml are properly set according to your environment.


## Deploy Grafana component (optional)

This component visualizes the predicted metrics. This is optional and can be ignored if you do not need it.
Alameda core function can work without this Grafana component. Please follow the following commands as your needs.

```
$ cd <alameda>/example/deployment/kubernetes
$ kubectl create -f 3rdParty/grafana.yaml
``` 

