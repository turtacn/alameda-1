# What is Alameda

Alameda is a prediction engine that foresees future resource usage of your Kubernetes cluster from the cloud layer down to the pod level. We use machine learning technology to provide intelligence that enables dynamic scaling and scheduling of your containers - effectively making us the “brain” of Kubernetes resource orchestration. By providing full foresight of resource availability, demand, health, impact and SLA, we enable cloud strategies that involve changing provisioned resources in real time. 

For more details, please refer to https://github.com/containers-ai/alameda

# Alameda deployment with Helm chart

> **Note**: To deploy Alameda by Helm charts, please install [Helm](https://docs.helm.sh/using_helm/#quickstart-guide) first.

According to Alameda [design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), it is composed of several components. You can find their Helm charts in the respective subfolders. 

- Charts of the following components are located at `./alameda`
  - operator
  - alameda-ai
  - datahub
- Prometheus chart is located at `./prometheus`
- InfluxDB chart is located at `./influxdb`
- Grafana chart is located at `./grafana`. Alameda also provides customized dashboard json files at `./grafana/dashboards/` and they will be imported when Grafana chart is deployed.

To get Alameda running, *operator*, *alameda-ai*, *datahub*, *Prometheus* and *InfluxDB* must be deployed.
> **Note**: You can levarage the running *Prometheus*, *InfluxDB* and *Grafana* instances in your cluster. Please refer to ./alameda/values.yaml to configure the connections between Alameda and these components.

## TL;DR;

```console
$ git clone https://github.com/containers-ai/alameda
$ helm install stable/influxdb --version 1.1.0 --name influxdb --namespace monitoring
$ helm install stable/prometheus-operator --version 4.3.1 --name prometheus --namespace monitoring -f ./prometheus-operator/values.yaml
$ helm install --name alameda --namespace alameda ./alameda
$ helm install --name grafana --namespace monitoring ./grafana/
```

