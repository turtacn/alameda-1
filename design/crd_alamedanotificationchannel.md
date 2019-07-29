## AlamedaNotificationChannel Custom Resource Definition

`notifier` uses `AlamedaNotificationChannel` to send events.

Here is an example _alamedanotificationchannel_ CR:

```
  apiVersion: notifying.containers.ai/v1alpha1
  kind: AlamedaNotificationChannel
  metadata:
    annotations:
      notifying.containers.ai/test-channel: start
      notifying.containers.ai/test-channel-to: to@example.com
    name: alamedanotificationchannel-sample
  spec:
    type: email
    email:
      server: mail.example.com
      port: 465
      from: from@example.com
      username: username
      password: password
      encryption: tls
```

### Test Channl Annotation

  Add annotation `notifying.containers.ai/test-channel: start` to `AlamedaNotificationChannel` CR,
  `notifier` test configuration immediately. Once channel test finished, annotation will be updated
  to `notifying.containers.ai/test-channel: done`.
  To test email type channel, add the annotation `notifying.containers.ai/test-channel-to: <recipient email>`
  to specify the recipient.

## Schema of AlamedaNotificationChannel

- Field: metadata
  - type: ObjectMeta
  - description: This follows the ObjectMeta definition in [Kubernetes API Reference](https://kubernetes.io/docs/reference/#api-reference)
- Field: spec
  - type: [AlamedaNotificationChannelSpec](#alamedanotificationchannelspec)
  - description: Spec of AlamedaNotificationChannel

### AlamedaNotificationChannelSpec

- Field: type
  - type: string
  - description: channel type
- Field: email
  - type: [AlamedaEmail](#alamedaemail)
  - description: email server configuration

### AlamedaEmail

- Field: server
  - type: string
  - description: the mail server host
- Field: port
  - type: int
  - description: the port of mail server
- Field: username
  - type: string
  - description: username used to login mail server
- Field: password
  - type: string
  - description: password used to login mail server
- Field: encryption
  - type: string
  - description: encryption of mail communication channel, the default value is tls
