## AlamedaScaler Custom Resource Definition

After Alameda is installed, it does not orchestrate any pod resources by default.
Alameda use `alamedascaler` [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) as a channel for users to tell Alameda which pods needs autoscaling services and what policy to follow.

Here is an example of `alamedascaler` CR:

```
  apiVersion: autoscaling.containers.ai/v1alpha1
  kind: AlamedaScaler
  metadata:
    name: alameda
    namespace: webapp
  spec:
    policy: stable
    selector:
      matchLabels:
        app: nginx
```

In this example, it creates an `AlamedaScaler` CR with name `alameda` in namespace `webapp`. With this CR, Alameda will look for K8s api objects with label `app` equals to `nginx` of the same `webapp` namespace. Any containers derivated from the found objects will be managed for their resource usages by Alameda. The `policy` field also instructs Alameda to recommend resource usage with `stable` policy.

When `AlamedaScaler` CR is created in K8s, Alameda will process it and add selected pods information to it. For example, the above `AlamedaScaler` CR example will have the following contents when it is got from K8s with yaml format.

```
apiVersion: v1
items:
- apiVersion: autoscaling.containers.ai/v1alpha1
  kind: AlamedaScaler
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"autoscaling.containers.ai/v1alpha1","kind":"AlamedaScaler","metadata":{"annotations":{},"name":"as","namespace":"alameda"},"spec":{"enable":true,"policy":"stable","selector":{"matchLabels":{"app.kubernetes.io/name":"alameda-ai"}}}}
    creationTimestamp: "2019-03-05T05:51:34Z"
    generation: 2
    name: as
    namespace: alameda
    resourceVersion: "1232719"
    selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/alameda/alamedascalers/as
    uid: bb9e1b3f-3f0a-11e9-b062-08606e0a1cbb
  spec:
    enable: true
    policy: stable
    selector:
      matchLabels:
        app.kubernetes.io/name: alameda-ai
  status:
    alamedaController:
      deploymentconfigs: {}
      deployments:
        alameda/alameda-ai:
          name: alameda-ai
          namespace: alameda
          pods:
            alameda/alameda-ai-7f5b6b6d8-8fqrv:
              containers:
              - name: alameda-ai
                resources: {}
              name: alameda-ai-7f5b6b6d8-8fqrv
              namespace: alameda
              uid: 2eb43d4c-3eee-11e9-b062-08606e0a1cbb
          uid: 28c96445-39b7-11e9-b062-08606e0a1cbb
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""

```

The `status` field shows no `deploymentconfigs` is selected and one `deployment` called `alameda-ai` is seleted.

When an `AlamedaScaler` CR is created in K8s, Alameda will also create `AlamedaRecommendation` CR(s) for each selected pod to expose resource recommendations. For example, in the above example, users can see a `AlamedaRecommendation` CR called `alameda-ai-7f5b6b6d8-8fqrv` is created. Here you can find more information about [`AlamedaRecommendation` CRD](./crd_alamedarecommendation.md).
```
$ kubectl get alamedarecommendations -n alameda
NAME                         AGE
alameda-ai-7f5b6b6d8-8fqrv   18m

```

## Current configurable settings

- Current supported K8s api objects are:
  - ```Deployment``` of api ```apps/v1```, and
  - ```DeploymentConfig``` of api ```apps.openshift.io/v1```

  Alameda will automatically look up those supported K8s api objects of the matched label.

- The supported policies are:
  - `stable`, and
  - `compact` (cost-saving).

- And Alameda only process the `matchLabels` field of `selector`.

