# Before deploy 

Before deploying components, execute following command.
```
$ oc adm policy add-scc-to-user privileged system:serviceaccount:openshift-monitoring:smart-exporter
``` 

# Components to deploy

InfluxDB: Remote data storage of Prometheus.In InfluxDB v1.5 and earlier, all Prometheus data goes into a single measurement named _, and the Prometheus measurement name is stored in the __name__ label.In InfluxDB v1.6+, every Prometheus measurement gets its own InfluxDB measurement.

# Components need updated

Prometheus: To write metrics to InfluxDB, add configuration below into Prometheus's spec which Prometheus Operator ensures that a deployment matching the resource definition.

You may need to get the name of Prometheus which needs updating by the following command.

```
$ oc -n openshift-monitoring get prometheus
```

Configuration leverages Prometheus to write metrics to InfluxDB.

```
spec:
  remoteWrite:
  - url: http://influxdb:8086/api/v1/prom/write?db=prometheus
    basicAuth:
      username:
        name: influxdb-admin-auth
        key: username
      password:
        name: influxdb-admin-auth
        key: password
```