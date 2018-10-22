# Before deploy 

Before deploying components, execute following command.
```
$ oc create namespace alameda-monitoring
$ oc adm policy add-scc-to-user privileged system:serviceaccount:alameda-monitoring:smart-exporter
$ oc adm policy add-scc-to-user privileged system:serviceaccount:alameda-monitoring:node-exporter
```

# Components to deploy

Prometheus: Collect metrics from k8s cAdvisor and pods. Via service discovery, Promethesu will scrape and keep targets from pods with the annotations below. You can also merge the prometheus's configuration of alameda into existing Prometheus.Inspect comments in ConfigMap "prometheus" to see how to properly configure Prometheus to scrape targets Alameda required.

```
annotations:
    alameda/scrape: "true"
```

Smart-exporter:Collect smart data from host via smartcmontools.

Node-exporter:Prometheus exporter for hardware and OS metrics exposed by *NIX kernels, written in Go with pluggable metric collectors.

Influxdb: Remote data storage of Prometheus.In InfluxDB v1.5 and earlier, all Prometheus data goes into a single measurement named _, and the Prometheus measurement name is stored in the __name__ label.In InfluxDB v1.6+, every Prometheus measurement gets its own InfluxDB measurement.

