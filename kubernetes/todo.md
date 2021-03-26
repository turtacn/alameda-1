# 数据对齐

## metrics set

|指标集|状态|
|--|--|
|CoreDNS|ok|
|etcd|nodata|
|Kubernetes/API server|ok|
|Kubernetes/Compute Resources/Cluster|partial|
|Kubernetes/Compute Resources/Namespace (Pods)|partial|
|Kubernetes/Compute Resources/Namespace (Workloads)|partial|
|Kubernetes/Compute Resources/Node (Pods)|ok|
|Kubernetes/Compute Resources/Pod|ok|
|Kubernetes/Compute Resources/Workload|ok|
|Kubernetes/Controller Manager|ok|
|Kubernetes/Kubelet|ok|
|Kubernetes/Networking/Cluster|ok|
|Kubernetes/Networking/Namespace (Pods)|ok|
|Kubernetes/Networking/Namespace (Workload)|ok|
|Kubernetes/Networking/Pod|ok|
|Kubernetes/Networking/Workload|ok|
|Kubernetes/Persistent Volumes|ok|
|Kubernetes/Proxy|nodata|
|Kubernetes/Scheduler|ok|
|Kubernetes/StatefulSets|ok|
|Nodes|ok|
|Prometheus Overview|ok|
|Prometheus Stats|partial|
|USE Method/Cluster|ok|
|USE Method/Node|ok|

## 缺失的指标

|指标|处理方法|详情|
|--|--|--| 
|instance:node_cpu_utilisation:rate1m|[指标替换](https://github.com/prometheus/node_exporter/issues/1454)|node:node_cpu_utilisation:avg1m * node:node_num_cpu:sum / scalar(sum(node:node_num_cpu:sum))|
|node:node_memory_utilisation:ratio| | instance:node_memory_utilisation:ratio|
|node:node_disk_utilisation:avg_irate|||