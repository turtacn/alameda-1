# Before deploy 

Before deploying components, execute following command.
```
$ oc create namespace alameda-monitoring
$ oc adm policy add-scc-to-user privileged system:serviceaccount:alameda-monitoring:smart-exporter
$ oc adm policy add-scc-to-user privileged system:serviceaccount:alameda-monitoring:node-exporter
```

# Components to deploy

Prometheus: Collect metrics from k8s cAdvisor and pods.Via service discovery, Prometheus will scrape metrics from pods with the annotations below.You can also merge the prometheus's configuration of alameda into existing Prometheus.There are two scrape_configs in Prometheus's configuration file that Alameda required.

First, job name "exporter", defining that Prometheus will create a target for each container's exposed ports in pods, fetch metrics from endpoint "/metrics" in "http" protocol and keep results that pod of target has annotation "alameda/scrape" with value "true".

Second, job name "kubernetes-cAdvisor", similar to first scrape config, dynamically create targets to scrape from each nodes in kubernetes, those targets will fetch metrics from kubernetes api server's endpoint "/api/v1/nodes/${1}/proxy/metrics/cadvisor", which ${1} in endpoints will be replaced with node's name, in "https" protocol.

For more configuration's detail, please inspect comments in ConfigMap "prometheus".

```
annotations:
    alameda/scrape: "true"
```

Smart-Exporter: Collect smart data from host via smartcmontools.

Node-Exporter: Prometheus exporter for hardware and OS metrics exposed by *NIX kernels, written in Go with pluggable metric collectors.

InfluxDB: Remote data storage of Prometheus.In InfluxDB v1.5 and earlier, all Prometheus data goes into a single measurement named _, and the Prometheus measurement name is stored in the __name__ label.In InfluxDB v1.6+, every Prometheus measurement gets its own InfluxDB measurement.

