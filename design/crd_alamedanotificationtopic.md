## AlamedaNotificationTopic Custom Resource Definition

`notifier` helps notify users event they are intested in.
You add topics in `AlamedaNotificationTopic` spec, `notifier`
sends events with matched topics setting via user provides channels.

Here is an example _alamedanotificationtopic_ CR:

```
  apiVersion: notifying.containers.ai/v1alpha1
  kind: AlamedaNotificationTopic
  metadata:
    name: alamedanotificationtopic-sample
  spec:
    channel:
      emails:
        - name: alamedanotificationchannel-sample
          to:
            - to@example.com
          cc:
            - cc@example.com
    topics:
      - type:
          - PodRegister
        subject:
          - namespace: default
            name: example
            kind: Pod
        level:
          - info
        source:
          - component: alameda-operator
```

## Schema of AlamedaNotificationTopic

- Field: metadata
  - type: ObjectMeta
  - description: This follows the ObjectMeta definition in [Kubernetes API Reference](https://kubernetes.io/docs/reference/#api-reference)
- Field: spec
  - type: [AlamedaNotificationTopicSpec](#alamedanotificationtopicspec)
  - description: Spec of AlamedaNotificationTopic

### AlamedaNotificationTopicSpec

- Field: disabled
  - type: bool
  - description: disable topic notification
- Field: topics
  - type: [AlamedaTopic](#alamedatopic) array
  - description: subscribe topics to notify
- Field: channel
  - type: [AlamedaEmailChannel](#alamedaemailchannel)
  - description: notify events via channel

### AlamedaTopic

- Field: type
  - type: string array
  - description: event types need to be notified
- Field: subject
  - type: [AlamedaSubject](#alamedasubject) array
  - description: event subjects need to be notified
- Field: level
  - type: string array
  - description: event levels need to be notified
- Field: source
  - type: [AlamedaSource](#alamedasource) array
  - description: event sources need to be notified

### AlamedaSource
- Field: host
  - type: string
  - description: source host
- Field: component
  - type: string
  - description: source component

### AlamedaSubject

- Field: kind
  - type: string
  - description: kubernetes resource kind
- Field: namespace
  - type: string
  - description: kubernetes resource namespace
- Field: name
  - type: string
  - description: kubernetes resource name
- Field: apiVersion
  - type: string
  - description: kubernetes resource API version

### AlamedaChannel

- Field: emails
  - type [AlamedaEmailChannel](#alamedaemailchannel) array
  - description: email notification channel to used and email header information

### AlamedaEmailChannel

- Field: name
  - type: string
  - description: email channel name
- Field: to
  - type: string array
  - description: email recipients
- Field: cc
  - type: string array
  - description: email recipients in cc list
