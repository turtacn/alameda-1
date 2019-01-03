For more details, please goto https://github.com/helm/charts/tree/master/stable/grafana

# Grafana Helm Chart

* Installs the web dashboarding system [Grafana](http://grafana.org/)

## TL;DR;

```console
$ helm install stable/grafana --name grafana --namespace monitoring -f values.yaml
```

Compared to the upstream `values.yaml`, the supplied `values.yaml` add a datasource configuration:  
```
datasources:
  datasources.yaml:
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus-server
      access: proxy
      isDefault: true
```
