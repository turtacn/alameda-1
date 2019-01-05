# Alameda Helm Chart

* Installs the brain of Kubernetes resource orchestration [Alameda](https://github.com/containers-ai/alameda)

## Alameda components
- operator
- alameda-ai
- datahub
- Prometheus
- InfluxDB
- Grafana (optional)

In this Helm chart, *operator*, *alameda-ai*, *datahub* and *Grafana*(if enabled in values.yaml) will be deployed and connection to Prometheus and InfluxDB will be configured.

## Requirements

Alameda levarages Prometheus to collect metrics and InfluxDB to store predictions with the following requirements.

1. Alameda requires a running Prometheus with following data exporters and record rules:  
  - Data Exporters  
    - cAdvisor
  - Recording Rules:   
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
2. Alameda requires a running InfluxDB. It will create database *cluster_status*, *recommendation* and *prediction* if they does not exist.

## TL;DR;

```console
$ git clone https://github.com/containers-ai/alameda
$ cd alameda/helm/alameda
$ helm install --name alameda --namespace alameda .
```
It will deploy Alameda in *alameda* namespace with *alameda* release name.

## Installing the Chart

To install the chart into `my-namespace` namespace with the release name `my-release`:

```console
$ helm install --name my-release --namespace my-namespace .
```

## Uninstalling the Chart

To uninstall/delete the my-release deployment:

```console
$ helm delete my-release --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.


