
# List of Metrics Exported by alameda-ai-dispatcher

- metric name: alameda_ai_dispatcher_container_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target container. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which container it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target container
    - pod_name: The pod where the target container belongs to
    - pod_namespace: The pod namespace where the target container belongs to

- metric name: alameda_ai_dispatcher_container_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target container. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which container it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target container
    - pod_name: The pod where the target container belongs to
    - pod_namespace: The pod namespace where the target container belongs to

- metric name: alameda_ai_dispatcher_pod_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a pod encountered. Since a pod may contain multiple containers and a container may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which pod it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target pod
    - namespace: The namespace where the target pod belongs to

- metric name: alameda_ai_dispatcher_pod_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which pod it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target pod
    - namespace: The namespace where the target pod belongs to

- metric name: alameda_ai_dispatcher_pod_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far on waiting for remodeling requests being completed
  - labels: The following labels are used to tag this metric that which pod it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target pod
    - namespace: The namespace where the target pod belongs to

- metric name: alameda_ai_dispatcher_node_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target node. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which node it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target node

- metric name: alameda_ai_dispatcher_node_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target node. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which node it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target node

- metric name: alameda_ai_dispatcher_node_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a node encountered. Since a node may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which node it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target node

- metric name: alameda_ai_dispatcher_node_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which node it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target node

- metric name: alameda_ai_dispatcher_node_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which node it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target node

- metric name: alameda_ai_dispatcher_gpu_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target gpu. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which gpu it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **DUTY_CYCLE** or **MEMORY_USAGE_BYTES**
    - host: The host of the target gpu
    - minor_number: The target gpu id of the host

- metric name: alameda_ai_dispatcher_gpu_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target gpu. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which gpu it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **DUTY_CYCLE** or **MEMORY_USAGE_BYTES**
    - host: The host of the target gpu
    - minor_number: The target gpu id of the host

- metric name: alameda_ai_dispatcher_gpu_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a gpu encountered. Since a gpu may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which gpu it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - host: The host of the target gpu
    - minor_number: The target gpu id of the host

- metric name: alameda_ai_dispatcher_gpu_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which gpu it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - host: The host of the target gpu
    - minor_number: The target gpu id of the host

- metric name: alameda_ai_dispatcher_gpu_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which gpu it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - host: The host of the target gpu
    - minor_number: The target gpu id of the host

- metric name: alameda_ai_dispatcher_namespace_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target namespace. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which namespace it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target namespace

- metric name: alameda_ai_dispatcher_namespace_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target namespace. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which namespace it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target namespace

- metric name: alameda_ai_dispatcher_namespace_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a namespace encountered. Since a namespace may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which namespace it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target namespace

- metric name: alameda_ai_dispatcher_namespace_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which namespace it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target namespace

- metric name: alameda_ai_dispatcher_namespace_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which namespace it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target namespace

- metric name: alameda_ai_dispatcher_application_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target application. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which application it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target application
    - namespace: The namespace where the target application belongs to

- metric name: alameda_ai_dispatcher_application_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target application. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which application it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target application
    - namespace: The namespace where the target application belongs to

- metric name: alameda_ai_dispatcher_application_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a application encountered. Since a application may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which application it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target application
    - namespace: The namespace where the target application belongs to

- metric name: alameda_ai_dispatcher_application_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which application it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target application
    - namespace: The namespace where the target application belongs to

- metric name: alameda_ai_dispatcher_application_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which application it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target application
    - namespace: The namespace where the target application belongs to

- metric name: alameda_ai_dispatcher_cluster_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target cluster. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which cluster it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target cluster

- metric name: alameda_ai_dispatcher_cluster_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target cluster. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which cluster it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target cluster

- metric name: alameda_ai_dispatcher_cluster_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a cluster encountered. Since a cluster may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which cluster it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target cluster

- metric name: alameda_ai_dispatcher_cluster_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which cluster it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target cluster

- metric name: alameda_ai_dispatcher_cluster_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which cluster it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target cluster

- metric name: alameda_ai_dispatcher_controller_metric_mape
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target controller. The accuracy measurement used in this metric is MAPE. As time passes, a drifted model with increasing MAPE is expected
  - labels: The following labels are used to tag this metric that which controller it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target controller
    - namespace: The namespace where the target controller belongs to
    - kind: The kind of the target controller

- metric name: alameda_ai_dispatcher_controller_metric_rmse
  - type: gauge
  - description: This metric shows the accuracy of the latest model that models the behavior of the target controller. The accuracy measurement used in this metric is RMSE. As time passes, a drifted model with increasing RMSE is expected
  - labels: The following labels are used to tag this metric that which controller it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - metric_type: The type of workload this metric is reporting can be **CPU_USAGE_SECONDS_PERCENTAGE** or **MEMORY_USAGE_BYTES**
    - name: The name of the target controller
    - namespace: The namespace where the target controller belongs to
    - kind: The kind of the target controller

- metric name: alameda_ai_dispatcher_controller_metric_drift_total
  - type: counter
  - description: This metric shows how many times drifts that a controller encountered. Since a controller may have multiple workloads to model, we increase the counting number once any of the above models are drifted
  - labels: The following labels are used to tag this metric that which controller it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target controller
    - namespace: The namespace where the target controller belongs to
    - kind: The kind of the target controller

- metric name: alameda_ai_dispatcher_controller_model_seconds
  - type: gauge
  - description: This metric shows how many seconds have lasted since a remodeling request is sent until the job is completed
  - labels: The following labels are used to tag this metric that which controller it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target controller
    - namespace: The namespace where the target controller belongs to
    - kind: The kind of the target controller

- metric name: alameda_ai_dispatcher_controller_model_seconds_total
  - type: counter
  - description: This metric shows how many seconds have been spent so far to wait for remodeling requests are completed
  - labels: The following labels are used to tag this metric that which controller it is reporting
    - data_granularity: The granularity can be **30s**, **1h**, **6h**, or **24h**
    - name: The name of the target controller
    - namespace: The namespace where the target controller belongs to
    - kind: The kind of the target controller
