# What is Federator.ai

Please refer to https://github.com/containers-ai/alameda

# Federator.ai deployment with Helm

> **Note**: To deploy Federator.ai by Helm charts, please install [Helm](https://docs.helm.sh/using_helm/#quickstart-guide) first.

According to Federator.ai [design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), it is composed of several components. You can find their Helm charts in the respective subfolders. 

- Chart of the following components are located at [./alameda](./alameda)
  - operator
  - alameda-ai
  - datahub
- Prometheus chart is located at [./prometheus](./prometheus) 
- InfluxDB chart is located at [./influxdb](./influxdb)
- Grafana chart is located at [./grafana](./grafana)

To get Federator.ai running, *operator*, *alameda-ai*, *datahub*, *Prometheus* and *InfluxDB* must be deployed.
> **Note**: You can levarage the running *Prometheus*, *InfluxDB* and *Grafana* instances in your cluster. Please refer to ./alameda/values.yaml to configure the connections.

## TL;DR;

```console
$ git clone https://github.com/containers-ai/alameda
$ helm install stable/influxdb --name influxdb --namespace monitoring
$ helm install stable/prometheus --name prometheus --namespace monitoring -f ./prometheus/values.yaml
$ helm install --name alameda --namespace alameda ./alameda
$ helm install stable/grafana --name grafana --namespace monitoring -f ./grafana/values.yaml
```

