For more details, please goto https://github.com/helm/charts/tree/master/stable/prometheus

# Prometheus

[Prometheus](https://prometheus.io/), a [Cloud Native Computing Foundation](https://cncf.io/) project, is a systems and service monitoring system. It collects metrics from configured targets at given intervals, evaluates rule expressions, displays the results, and can trigger alerts if some condition is observed to be true.

## TL;DR;

```console
$ helm install stable/prometheus --name prometheus --namespace monitoring -f values.yaml
```

Compared to the upstream `values.yaml`, the supplied `values.yaml` adds the following recording rules:  
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
