## AlamedaRecommendation Custom Resource Definition

`AlamedaRecommendation` CRD is one of the ways to expose Alameda recommendation results. When users create [`AlamedaScaler` CRs](./crd_alamedascaler.md), Alameda will create `AlamedaRecommendation` CRs for all the seleted pods. Users can get it by calling K8s api calls.

> **Note**: Currently there are three ways to get Alameda recommendation results:
- `AlamedaRecommendation` CRs. Users can get it by calling K8s api calls.
- Visualized with Grafana dashboards. The data is pulled directly from influxDB.
- gRPC API calls to `datahub` component. `Datahub` component runs as a gRPC server to handle all the data access needed in Alameda.

Here is an example of `alamedarecommendation` CR:

```
apiVersion: v1
items:
- apiVersion: autoscaling.containers.ai/v1alpha1
  kind: AlamedaRecommendation
  metadata:
    creationTimestamp: "2019-03-05T05:51:34Z"
    generation: 19
    labels:
      alamedascaler: as.alameda
    name: alameda-ai-7f5b6b6d8-8fqrv
    namespace: alameda
    ownerReferences:
    - apiVersion: autoscaling.containers.ai/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: AlamedaScaler
      name: as
      uid: bb9e1b3f-3f0a-11e9-b062-08606e0a1cbb
    resourceVersion: "1234985"
    selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/alameda/alamedarecommendations/alameda-ai-7f5b6b6d8-8fqrv
    uid: bba55ef6-3f0a-11e9-b062-08606e0a1cbb
  spec:
    containers:
    - name: alameda-ai
      resources:
        limits:
          cpu: 1654m
          memory: 388136Ki
        requests:
          cpu: 1649m
          memory: 388136Ki
  status: {}
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

In this example, this `AlamedaRecommendation` CR responds to `AlamedaScaler` *as* of namespace *alameda* and is created for pod `alameda-ai-7f5b6b6d8-8fqrv`. The recommendations for this pod is to set cpu request and limit to 1649m and 1654m, and to set memory request and limit to 388136Ki and 388136Ki.

> **Note:**: The recommendations will be updated when new recommendations are available.

