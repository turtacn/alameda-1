# QuickStart

This document guides you from build, deploy to use Federator.ai.
- Build Federator.ai
- Deploy Federator.ai
- Use Federator.ai 

## Build Federator.ai
Please refer to the [build](./build.md) guide.

## Deploy Federator.ai
Please refer to the [deploy](./deploy.md) guide.

## Use Federator.ai

### Specify a target to request Federator.ai services
User can create a custom resource of *AlamedaScaler* custom resource definition (CRD) to instruct Federator.ai that
1. which Pod(s) to watch by Kubernetes *selector* construct, and
2. what policy that Federator.ai should use to give recommendations.

Currently Federator.ai provides *stable* and *compact* policy. The following is an example to instruct Federator.ai to watch Pod(s) at *webapp* namespace with *nginx* label and *stable* policy.
```
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaScaler
metadata:
  name: alameda
  namespace: webapp
spec:
  policy: stable
  enable: true
  selector:
    matchLabels:
      app: nginx
```

You can list all Pods that are watched by Federator.ai with:
```
$ kubectl get alamedascalers --all-namespaces
```

### Retrieve Federator.ai prediction and recommendation result
Federator.ai outputs raw workload recommendations in a global planning manner for all the pods watched by Federator.ai.
They are presented as *alamedarecommendation* CRD.
You can check Federator.ai recommendation results by:
```
$ kubectl get alamedarecommendation --all-namespaces
```

### Example

- Deploy a nginx application example by:
    ```
    $ cd <alameda>/example/samples/nginx
    $ kubectl create -f nginx_deployment.yaml
    ```
- Request Federator.ai to predict and recommend the resource usage for nginx Pods by:
    ```
    $ cd <alameda>/example/samples/nginx
    $ kubectl create -f alamedascaler.yaml
    ```
You can check that Federator.ai is watching the nginx Pods by:
```
$ kubectl get alamedascaler --all-namespaces
NAMESPACE   NAME      AGE
webapp      alameda   5h
$ kubectl get alamedascaler alameda -n webapp -o yaml
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaScaler
metadata:
  annotations:
    annotation-k8s-controller: |-
      {
        "DeploymentMap": {
          "2e5055f0-f47e-11e8-8913-88d7f6561288": {
            "UID": "2e5055f0-f47e-11e8-8913-88d7f6561288",
            "Name": "nginx-deployment",
            "PodMap": {
              "74d476c5-f48b-11e8-8913-88d7f6561288": {
                "UID": "74d476c5-f48b-11e8-8913-88d7f6561288",
                "Name": "nginx-deployment-88644cb8c-clwrx",
                "Containers": [
                  {
                    "Name": "nginx"
                  }
                ]
              },
              "cc73a788-f487-11e8-8913-88d7f6561288": {
                "UID": "cc73a788-f487-11e8-8913-88d7f6561288",
                "Name": "nginx-deployment-88644cb8c-xbdgj",
                "Containers": [
                  {
                    "Name": "nginx"
                  }
                ]
              }
            }
          }
        }
      }
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"autoscaling.containers.ai/v1alpha1","kind":"AlamedaScaler","metadata":{"annotations":{},"name":"alameda","namespace":"alameda"},"spec":{"enable":true,"policy":"compact","selector":{"matchLabels":{"app":"nginx"}}}}
  creationTimestamp: 2018-11-30T08:59:28Z
  generation: 1
  name: alameda
  namespace: webapp
  resourceVersion: "67988"
  selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/webapp/alamedascaler/alameda
  uid: 3e8ea661-f47e-11e8-8913-88d7f6561288
spec:
  enable: true
  policy: compact
  selector:
    matchLabels:
      app: nginx
status: {}
```
And the prediction and recommendation for the nginx Pods are:
```
$ kubectl get alamedarecommendation alameda -n webapp -o yaml
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaResourcePrediction
metadata:
  creationTimestamp: 2018-11-30T08:59:28Z
  generation: 1
  name: alameda
  namespace: webapp
  resourceVersion: "70278"
  selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/webapp/alamedarecommendation/alameda
  uid: 3e9dc05e-f47e-11e8-8913-88d7f6561288
spec:
  selector:
    matchLabels:
      app: nginx
status:
  prediction:
    Deployments:
      2e5055f0-f47e-11e8-8913-88d7f6561288:
        Name: alameda
        Namespace: alameda
        Pods:
          74d476c5-f48b-11e8-8913-88d7f6561288:
            Containers:
              nginx:
                InitialResource:
                  limits:
                    memory: "200Mi"
                  requests:
                    memory: "100Mi"
                Name: nginx
                Recommendations:
                - Date: 2018-11-30 10:43:30 +0000 UTC
                  Resources:
                    limits:
                      memory: "200Mi"
                    requests:
                      memory: "100Mi"
                  Time: 1543574610
                - Date: 2018-11-30 10:46:30 +0000 UTC
                  Resources:
                    limits:
                      memory: "200Mi"
                    requests:
                      memory: "100Mi"
                  Time: 1543574790
                - Date: 2018-11-30 10:49:30 +0000 UTC
                  Resources:
                    limits:
                      memory: "200Mi"
                    requests:
                      memory: "100Mi"
                  Time: 1543574970
            Name: nginx-deployment-88644cb8c-clwrx
          cc73a788-f487-11e8-8913-88d7f6561288:
            Containers:
              nginx:
                InitialResource: {}
                Name: nginx
                RawPredict: {}
                Recommendations: []
            Name: nginx-deployment-88644cb8c-xbdgj
        UID: 2e5055f0-f47e-11e8-8913-88d7f6561288
```
