This chart installation is intended to install the influxdb chart of version 1.7 at https://kubernetes-charts.storage.googleapis.com. For more details, please goto https://github.com/helm/charts/tree/master/stable/influxdb

# InfluxDB

##  An Open-Source Time Series Database

[InfluxDB](https://github.com/influxdata/influxdb) is an open source time series database built by the folks over at [InfluxData](https://influxdata.com) with no external dependencies. It's useful for recording metrics, events, and performing analytics.

## QuickStart

```bash
$ helm install stable/influxdb --version 1.7 --name influxdb --namespace monitoring
```
