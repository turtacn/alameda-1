For more details, please goto https://github.com/helm/charts/tree/master/stable/prometheus

# Prometheus

[Prometheus](https://prometheus.io/), a [Cloud Native Computing Foundation](https://cncf.io/) project, is a systems and service monitoring system. It collects metrics from configured targets at given intervals, evaluates rule expressions, displays the results, and can trigger alerts if some condition is observed to be true.

## TL;DR;

```console
$ helm install stable/prometheus --name prometheus --namespace monitoring -f values.yaml
```

Compared to the upstream `values.yaml`, the supplied `values.yaml` adds a relabel target *kubernetes_pod* in job *kubernetes-service-endpoints* and the following recording rules:  
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
    - expr: 1 - avg by (kubernetes_node) (rate(node_cpu{job="kubernetes-service-endpoints",mode="idle"}[1m]) * on(kubernetes_namespace, kubernetes_pod) group_left(node) node_namespace_pod:kube_pod_info:)
      record: node:node_cpu_utilisation:avg1m
    - expr: node_memory_MemTotal_bytes
      record: node_memory_MemTotal
    - expr: node_memory_MemFree_bytes
      record: node_memory_MemFree
    - expr: node_memory_Cached_bytes
      record: node_memory_Cached
    - expr: node_memory_Buffers_bytes
      record: node_memory_Buffers
    - expr: sum
        by (kubernetes_node) ((node_memory_MemFree{job="kubernetes-service-endpoints"} + node_memory_Cached{job="kubernetes-service-endpoints"}
        + node_memory_Buffers{job="kubernetes-service-endpoints"}) * on(kubernetes_namespace, kubernetes_pod) group_left(kubernetes_node)
        node_namespace_pod:kube_pod_info:)
      record: node:node_memory_bytes_available:sum
    - expr: sum
        by(kubernetes_node) (node_memory_MemTotal{job="kubernetes-service-endpoints"} * on(kubernetes_namespace, kubernetes_pod)
        group_left(kubernetes_node) node_namespace_pod:kube_pod_info:)
      record: node:node_memory_bytes_total:sum
```
