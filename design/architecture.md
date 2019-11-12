## What is Alameda 

Alameda provides AI-driven resource orchestration intelligence for Kubernetes. More specifically, it provides the intelligence for autonomous balancing, scaling, and scheduling by using machine learning. Alameda learns the continuing changes of computing resources from Kubernetes clusters, predicts the future computing resources demands of pods, and generates resource configuration recommendation for intelligently orchestrating the underlying computing resources without manual configuration.

The following figure illustrates how Alameda works in K8s. Users specify Pods and configure Alameda with a preferred policy. Then Alameda exposes raw predicted metrics and also bundled recommendations for them. Any resource orchestrators can benefit from Alameda's output by reacting to these machine-learned intelligence.

![usecase](./img/usecase.png)

## Features

The primary purpose of Alameda is to recommend optimal computing resource configuration for Kubernetes by utilizing AI-powered prediction capability. With this, IT admins can leave one of the hardest problems of running Kubernetes to Alameda. Features of Alameda include:

- AI-driven resource management for CPU, memory, and disks  
    Alameda AI Engine generates loading data predictions for the future time. The AI Engine learns patterns from the historical performance metrics of each node and pods running on it. For example, it predicts CPU metrics of the next 24 hours in 1-hour interval. Besides future performance metrics, Alameda also detects disk health and predicts life expectancy based on a disk's S.M.A.R.T. value. With these predicted metrics and data, Alameda optimizes resource provisioning for Kubernetes.

- Integral scaling considerations  
    When managing Kubernetes clusters, IT admins usually need to take care of how much resources are required by each pod, when a pod needs to scale up/down replicas and when a cluster needs to scale up/down a node. Though several autoscalers such as [VPA](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler), [HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/), and [CA](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) are introduced in Kubernetes, they work reactively and independently. Even more, scaling outcomes might not be desirable due to contradictory settings by humans. With Alameda, all these scaling considerations are taken into one integrated global planning.

- Policy-driven optimization  
    Alameda provides *stable* and *cost-saving* policies for users to orchestrate resources. With the *stable* policy, users can expect more available resources reserved on each node, which could effectively reduce pod restarting due to insufficient resources for the pods. With *cost-saving* policy, users can expect less running nodes, which could effectively reduce operational spending. This policy-based optimization simplifies resource management complexity. 

- Well integrated into a Kubernetes cluster  
    Alameda exposes metrics predictions to a time-series database and operation recommendations via Kubernetes' [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). Anyone can leverage these to build up their extensions.

## Alameda Position in K8s Ecosystem

As showed in the following figure, Alameda is integrated into the Kubernetest ecosystem as the brain of resource orchestration and cooperates with Kubernetes orchestrators by providing metric predictions and resource configuration recommendations.

![position](./img/position.png)

## Alameda Architecture

As the following figure shows, Alameda works with several components:

![architecture](./img/architecture.png)

1. data sources: Alameda leverage open-source projects for data collection and itself does not implement data collection agents. 
  - _Prometheus_ and _data exporters_: [Prometheus](https://prometheus.io/) is a popular open-source monitoring tool with large number of users. Alameda leverages Prometheus and third-parta data exporters to get historical metrics for model training.
  - _weavescope agent_: Alameda implements a wrapper to pull data from [Weave Scope](https://github.com/weaveworks/scope) for knowing the maps of applications as well as entire infrastructure.
2. _alameda-operator_: The Alameda operator introduces a CRD call *alamedascaler* to provide a channel for users to select Pods for requesting Alameda service. It will register those selected Pods and cluster nodes to alameda-ai for predictions and recommendations.
3. prediction: Considering the characteristics of machine learning computation, Alameda implements _alameda-ai_ component to work on model training and workload prediction jobs and _ai-dispatcher_ to dispatch jobs. _alameda-ai_ component replica is scaled based on how heavy the loading is.
4. recommendation: Recommendations from Alameda can be splited into two components, _fedemeter_ and _recommender_, to tackle the Day 1 (deployment) and Day 2 (operations) tasks for Kubernetes clusters. These components make recommendations based on the predicted workloads.
5. _alameda-analyzer_: Besides prediction and recommendation, Alameda also provides features such as _anomaly detection_ and _event correlation_. Imaging a hundreds of thousands of pods cluster, these AI-powered techniques can help operation efficiency in fast fetecting issues and filtering out the irrelavant.
6. _InfluxDB_: This time-series DB is used as a storage backend of whole Alameda system. It stores data such as predicted metrics and recommendations.
7. _datahub_: datahub plays an API and data gateway betwen Alameda components. All data access such as Prometheus and InfluxDB is through this component. The component itself is designed to be stateless and can be scaled out when loading is heavy.
8. execution: The execution part is optional in Alameda. Users may leverage Alameda's predictions and recommendations and implement their owned execution.
9. _grafana_: Alameda leverages the open-source Grafana to visualize predictions and recommendations. This component is optional and Alameda dashboards are installed by default when it is deployed.
10. _notifier_: This component provides nofitications such as email to notifiy users for events.

The following message sequence chart demonstrates how Alameda normally works.

![workflow](./img/workflow.png)


