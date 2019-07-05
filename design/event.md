## Proposal
To provide user (administrator/developer/application) monitor/analyze occurred events in alameda running system, this document defines the schema of the event and provides some scenarios that events might occurred while alameda is running. For join into cloud ecosystem futher, reference [cloudevent](https://github.com/cloudevents/spec) provides an general guideline when designing alameda event schema.

## Schema of event

- Field: time
  - type: string
  - format: rfc3339 in nanoseconds
  - description: Timestamp of when the occurrence happened.
  - constraints:
    -  required
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
  - type: string
  - format: '^\/(?<application_name>[^\/]+)\/(?<component_name>[^\/]+)$'
  - description: Identifies which component of application did the event happened.
  - constraints:
    - required
  - example:
    - /alameda/operator
    - /alameda/evitioner
- Field: type
  - type: string
  - description: Type of this event.
  - constraints:
    - required
  - example:
    - PodRegister
    - PodEvict
- Field: subtype
  - type: string
  - description: Subtype of this event.
  - constraints:
    -  Optional
- Field: version
  - type: string
  - description: Version of this event.
  - constraints:
    - required
  - example:
    - v1
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
    -  required
- Field: namespace
  - type: string
  - description: This describes the namespace of the subject of the event in the context of the event producer.
  - constraints:
    -  optional
  - example:
    - webapp
- Field: subject
  - type: string
  - description: This describes the subject of the event in the context of the event producer.
  - constraints:
    -  required
  - example:
    - nginx-deployment-7948f76569-lgw26
- Field: reason
  - type: string
  - description: A short description of this event that is human readable.
  - constraints:
    -  required
- Field: message
  - type: string
  - description: A full description of this event that is human readable. 
  - constraints:
    -  required
- Field: data
  - type: string
  - description: The event payload. Schema of this payload depends on the source, type and version. 
  - constraints:
    -  optional

## Examples
Scenearios that events that might be created.

New AlamedaScaler created by user. 
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "namespace": "webapp",
    "source": "/alameda/operator",
    "type": "AlamedaScalerCreate",
    "version": "v1",
    "level": "info",
    "subject": "alamedascaler-sample",
    "reason": "AlamedaScaler created",
    "message": "AlamedaScaler created"
}
```
Alameda-Datahub receives pod registration.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": "/alameda/datahub",
    "type": "PodRegister",
    "version": "v1",
    "level": "debug",
    "namespace": "webapp",
    "subject": "nginx-deployment-7948f76569-lgw26",
    "reason": "Register pod",
    "message": "Register nginx-deployment-7948f76569-lgw26 pod"
}
```
Alameda-AI predicts pod prediction.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": "/alameda/alameda-ai",
    "type": "PodPredictionCreate",
    "version": "v1",
    "level": "info",
    "namespace": "webapp",
    "subject": "nginx-deployment-7948f76569-lgw26",
    "reason": "Create pod prediction",
    "message": "nginx-deployment-7948f76569-lgw26 pod prediction created"
}
```
Alameda-Recommendator produces pod resources recommendation.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": "/alameda/recommendator",
    "type": "VPArecommendationCreate",
    "version": "v1",
    "level": "info",
    "namespace": "webapp",
    "subject": "nginx-deployment-7948f76569-lgw26",
    "reason": "Create pod recommendation",
    "message": "nginx-deployment-7948f76569-lgw26 pod recommendation created"
}
```
Alameda-Evictioner decides to evict Pod.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": "/alameda/evictioner",
    "type": "PodEvict",
    "version": "v1",
    "level": "info",
    "namespace": "webapp",
    "subject": "nginx-deployment-7948f76569-lgw26",
    "reason": "Evict pod",
    "message": "nginx-deployment-7948f76569-lgw26 pod evicted because cpu variance exceeds threshold"
}
```
Alameda-Admission-Controller patchs recommendation to the new created pod.
```
{
    "time": "1562295600000000000",
    "id": "545bf15e-36d0-447f-b79a-e61723c1b854",
    "source": "/alameda/admission-controller",
    "type": "PodPatch",
    "version": "v1",
    "level": "info",
    "namespace": "webapp",
    "subject": "nginx-deployment-7948f76569-",
    "reason": "Patch pod",
    "message": "Patch resource recommendation to new created pod under nginx-deployment-7948f76569- replicaset"
}
```