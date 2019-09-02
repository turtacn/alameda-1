## Proposal
To provide user (administrator/developer/application) monitor/analyze occurred events in alameda running system, this document defines the schema of the event and provides some scenarios that events might occurred while alameda is running. Reference [cloudevent](https://github.com/cloudevents/spec) provides an general guideline when designing alameda event schema.

## Schema of event

- Field: time
  - type: string
  - format: Epoch time in nanoseconds
  - description: Timestamp of when the occurrence happened.
  - constraints:
    - required
- Field: id
  - type: string
  - format: uuid
  - description: Identifies the event. Event producer must ensure that “source” + “id” is unique for each distinct event. 
  - constraints:
    - required
- Field: clusterID
  - type: string
  - format: uuid
  - description: Identifies which cluster did the event occurred. This field should be provided when running alameda with cloud service to distiguish event from each individual user cluster. 
  - constraints:
    - optional
- Field: source
  - type: [eventSource](#eventsource)
  - description: Identifies which source did the event happened.
  - constraints:
    - required
- Field: type
  - type: string
  - enum:
    - AlamedaScalerCreate 
    - AlamedaScalerDelete
    - NodeRegister
    - DeploymentRegister  
    - DeploymentConfigRegister  
    - PodRegister  
    - NodeDeregister
    - DeploymentDeregister 
    - DeploymentConfigDeregister 
    - PodDeregister
    - NodePredictionCreate
    - PodPredictionCreate
    - VPARecommendationCreate
    - HPARecommendationCreate
    - PodEvict
    - PodPatch
    - AnomalyMetricDetect
    - AnomalyAnalysisCreate
    - ReplicasUpdate
  - description: Type of this event.
  - constraints:
    - required
- Field: version
  - type: string
  - enum:
    - v1
  - description: Version of this event.
  - constraints:
    - required
- Field: level
  - type: string
  - enum:
    - debug
    - info
    - warning
    - error
    - fatal
  - description: Level of this event.
  - constraints:
    - required
- Field: subject
  - type: [k8sObjectReference](#k8sobjectreference)
  - description: This describes the subject of the event in the context of the event producer.
  - constraints:
    - required
- Field: message
  - type: string
  - description: A description of this event that is human readable. 
  - constraints:
    - required
- Field: data
  - type: string
  - format: JSON object in string
  - description: The event payload. Schema of this payload depends on the source, type and version. 
  - constraints:
    -  optional

## EventSource
- Field: host
  - type: string
  - description: Identifies which host did the event happened.
  - constraints:
    -  optional
  - example:
    - node246110
- Field: component
  - type: string
  - description: Identifies which component of application did the event happened.
  - constraints:
    - required
  - example:
    - alameda-operator

## K8SObjectReference
- Field: kind
  - type: string
  - description: This describes the kind of the involved k8s resource of the event.
  - constraints:
    - required
  - example:
    - Pod
- Field: namespace
  - type: string
  - description: This describes the namespace of the involved k8s resource of the event.
  - constraints:
    -  optional
  - example:
    - webapp
- Field: name
  - type: string
  - description: This describes the name of the involved k8s resource of the event.
  - constraints:
    - required
  - example:
    - nginx-deployment-7948f76569-lgw26
- Field: apiVersion
  - type: string
  - description: This describes the apiVersion of the involved k8s resource of the event.
  - constraints:
    - required
  - example:
    - v1

## Examples
Scenarios that events that might be created.

New AlamedaScaler created by user. 
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "AlamedaScalerCreate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "AlamedaScaler",
        "namespace": "webapp",
        "name": "alamedascaler-sample",
        "apiVersion": "autoscaling.containers.ai/v1alpha1"
    },
    "message": "AlamedaScaler created"
}
```
AlamedaScaler deleated by user. 
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "AlamedaScalerDelete",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "AlamedaScaler",
        "namespace": "webapp",
        "name": "alamedascaler-sample",
        "apiVersion": "autoscaling.containers.ai/v1alpha1"
    },
    "message": "AlamedaScaler created"
}
```
Register node to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "NodeRegister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Node",
        "name": "node246100",
        "apiVersion": "v1"
    },
    "message": "Register node246100 node"
}
```
Register deployment to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "DeploymentRegister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Deployment",
        "namespace": "webapp",
        "name": "nginx-deployment",
        "apiVersion": "apps/v1"
    },
    "message": "Register nginx-deployment deployment"
}
```
Register deploymentConfig to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "DeploymentConfigRegister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "DeploymentConfig",
        "namespace": "webapp",
        "name": "nginx-deploymentConfig",
        "apiVersion": "apps.openshift.io/v1"
    },
    "message": "Register nginx-deploymentConfig deploymentConfig"
}
```
Register pod to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "PodRegister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Pod",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-lgw26",
        "apiVersion": "v1"
    },
    "message": "Register nginx-deployment-7948f76569-lgw26 pod"
}
```
Deregister node to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "NodeDeregister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Node",
        "name": "node246100",
        "apiVersion": "v1"
    },
    "message": "Deregister node246100 node"
}
```
Deregister deployment to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "DeploymentDeregister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Deployment",
        "namespace": "webapp",
        "name": "nginx-deployment",
        "apiVersion": "apps/v1"
    },
    "message": "Deregister nginx-deployment deployment"
}
```
Deregister deploymentConfig to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "DeploymentConfigDeregister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "DeploymentConfig",
        "namespace": "webapp",
        "name": "nginx-deploymentConfig",
        "apiVersion": "apps.openshift.io/v1"
    },
    "message": "Deregister nginx-deploymentConfig deploymentConfig"
}
```
Deregister pod to Alameda-Datahub.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-operator"
    },
    "type": "PodDeregister",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Pod",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-lgw26",
        "apiVersion": "v1"
    },
    "message": "Deregister nginx-deployment-7948f76569-lgw26 pod"
}
```
Alameda-AI predicts node prediction.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-ai"
    },
    "type": "NodePredictionCreate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Node",
        "name": "node246110",
        "apiVersion": "v1"
    },
    "message": "node246110 node prediction created"
}
```
Alameda-AI predicts pod prediction.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-ai"
    },
    "type": "PodPredictionCreate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Pod",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-lgw26",
        "apiVersion": "v1"
    },
    "message": "nginx-deployment-7948f76569-lgw26 pod prediction created"
}
```
Alameda-Recommendator produces pod resources recommendation.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-recommendator"
    },
    "type": "VPArecommendationCreate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Pod",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-lgw26",
        "apiVersion": "v1"
    },
    "message": "nginx-deployment-7948f76569-lgw26 pod recommendation created"
}
```
Alameda-Recommendator produces deployment resources recommendation.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-recommendator"
    },
    "type": "HPArecommendationCreate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Deployment",
        "namespace": "webapp",
        "name": "nginx-deployment",
        "apiVersion": "apps/v1"
    },
    "message": "nginx-deployment deployment recommendation created"
}
```
Alameda-Evictioner decides to evict Pod.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-evictioner"
    },
    "type": "PodEvict",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Pod",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-lgw26",
        "apiVersion": "v1"
    },
    "message": "nginx-deployment-7948f76569-lgw26 pod evicted because cpu variance exceeds threshold"
}
```
Alameda-Admission-Controller patchs recommendation to the new created pod.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "admission-controller"
    },
    "type": "PodPatch",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "ReplicaSet",
        "namespace": "webapp",
        "name": "nginx-deployment-7948f76569-",
        "apiVersion": "apps/v1"
    },
    "message": "Patch resource recommendation to new created pod under nginx-deployment-7948f76569- replicaset"
}
``` 
Alameda-Executor scales up/down controllers' replicas.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": {
        "component": "alameda-executor"
    },
    "type": "ReplicasUpdate",
    "version": "v1",
    "level": "info",
    "subject": {
        "kind": "Deployment",
        "namespace": "webapp",
        "name": "nginx-deployment",
        "apiVersion": "apps/v1"
    },
    "message": "Update Deployment webapp/nginx-deployment replicas from 2 to 4"
}
``` 