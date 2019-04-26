## AlamedaRecommendation Custom Resource Definition

_AlamedaRecommendation_ CRD is one of the three ways to expose Alameda recommendation results. When users create [`AlamedaScaler` CRs](./crd_alamedascaler.md), Alameda will create _AlamedaRecommendation_ CRs for each seleted pods. The main purpose of this CRD is to provide an integration point for programs (including Alameda itself) to leverage the Alameda outputs which are resource orchestration recommendations.

> **Note**: Currently there are three ways to see Alameda recommendation results:
- _AlamedaRecommendation_ CR. Users can get it by calling K8s api calls or watch it to get notification when the CR is updated.
- Visualized with Grafana dashboards. The data is pulled directly from InfluxDB.
- gRPC API calls to **datahub** component. **Datahub** component runs as a gRPC server to handle all the data access of Alameda including recommendation results.

Here is an example _alamedarecommendation_ CR:

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

In this example, this _AlamedaRecommendation_ CR responds to _AlamedaScaler_ _as_ in namespace _alameda_ and is created for pod _alameda-ai-7f5b6b6d8-8fqrv_. The recommendations for this pod is to set cpu request and limit to 1649m and 1654m, and to set memory request and limit to 388136Ki and 388136Ki.

> **Note:**: The recommendations will be updated when new recommendations are available.

## Schema of AlamedaRecommendation

- Field: metadata
  - type: ObjectMeta
  - description: This follows the ObjectMeta definition in [Kubernetes API Reference](https://kubernetes.io/docs/reference/#api-reference)
- Field: spec
  - type: [AlamedaRecommendationSpec](#alamedarecommendationspec)
  - description: Spec of AlamedaRecommendation

### AlamedaRecommendationSpec

- Field: containers
  - type: [ContainerResourceRecommendation](#containerresourcerecommendation) array
  - description: List of containers with resource recommendations.

### ContainerResourceRecommendation

- Field: name
  - type: string
  - description: name of container.
- Field: limits
  - type: object
  - description: Limits describes the **recommended** maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
- Field: requests
  - type: object
  - description: Requests describes the **recommended** minimum amount of compute resources required. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

