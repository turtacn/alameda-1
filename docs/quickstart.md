# QuickStart

This document guides you from build, deploy to use Alameda.
- Build Alameda
- Deploy Alameda
- Use Alameda

## Build Alameda
Please refer to the [build](./build.md) guide.

## Deploy Alameda
Please refer to the [deploy](./deploy.md) guide.

## Use Alameda

### Specify a target to request Alameda services
User can create a custom resource of *AlamedaResource* custom resource definition (CRD) to instruct Alameda that
1. which Pod(s) to watch by Kubernetes *selector* construct, and
2. what policy that Alameda should use to give recommendations.

Currently Alameda provides *stable* and *compact* policy. The following is an example to instruct Alameda to watch Pod(s) at *webapp* namespace with *nginx* label and *stable* policy.
```
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaResource
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

You can list all Pods that are watched by Alameda with:
```
$ kubectl get alamedaresources --all-namespaces
```

### Retrieve Alameda prediction and recommendation result
Alameda outputs raw workload metrics prediction and recommendations in a global planning manner for all the pods watched by Alameda.
They are presented as *alamedaresourceprediction* CRD.
You can check Alameda prediction and recommendation results by:
```
$ kubectl get alamedaresourceprediction --all-namespaces
```

### Example

- Deploy a nginx application example by:
    ```
    $ cd <alameda>/example/samples/nginx
    $ kubectl create -f nginx_deployment.yaml
    ```
- Request Alameda to predict and recommend the resource usage for nginx Pods by:
    ```
    $ cd <alameda>/example/samples/nginx
    $ kubectl create -f alamedaresource.yaml
    ```
You can check that Alameda is watching the nginx Pods by:
```
$ kubectl get alamedaresource --all-namespaces
NAMESPACE   NAME      AGE
webapp      alameda   5h
$ kubectl get alamedaresource alameda -n webapp -o yaml
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaResource
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
      {"apiVersion":"autoscaling.containers.ai/v1alpha1","kind":"AlamedaResource","metadata":{"annotations":{},"name":"alameda","namespace":"alameda"},"spec":{"enable":true,"policy":"compact","selector":{"matchLabels":{"app":"nginx"}}}}
  creationTimestamp: 2018-11-30T08:59:28Z
  generation: 1
  name: alameda
  namespace: webapp
  resourceVersion: "67988"
  selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/webapp/alamedaresources/alameda
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
$ kubectl get alamedaresourceprediction alameda -n webapp -o yaml
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaResourcePrediction
metadata:
  creationTimestamp: 2018-11-30T08:59:28Z
  generation: 1
  name: alameda
  namespace: webapp
  resourceVersion: "70278"
  selfLink: /apis/autoscaling.containers.ai/v1alpha1/namespaces/webapp/alamedaresourcepredictions/alameda
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
                RawPredict:
                  memory:
                    PredictData:
                    - Date: 2018-11-30 10:43:30 +0000 UTC
                      Time: 1543574610
                      Value: "0.037559220060725686"
                    - Date: 2018-11-30 10:44:00 +0000 UTC
                      Time: 1543574640
                      Value: "0.037559018003475955"
                    - Date: 2018-11-30 10:44:30 +0000 UTC
                      Time: 1543574670
                      Value: "0.03755901800000006"
                    - Date: 2018-11-30 10:45:00 +0000 UTC
                      Time: 1543574700
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:45:30 +0000 UTC
                      Time: 1543574730
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:46:00 +0000 UTC
                      Time: 1543574760
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:46:30 +0000 UTC
                      Time: 1543574790
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:47:00 +0000 UTC
                      Time: 1543574820
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:47:30 +0000 UTC
                      Time: 1543574850
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:48:00 +0000 UTC
                      Time: 1543574880
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:48:30 +0000 UTC
                      Time: 1543574910
                      Value: "0.03755901799999999"
                    - Date: 2018-11-30 10:49:00 +0000 UTC
                      Time: 1543574940
                      Value: "0.04930520163194117"
                    - Date: 2018-11-30 10:49:30 +0000 UTC
                      Time: 1543574970
                      Value: "0.03755922006414266"
                    - Date: 2018-11-30 10:50:00 +0000 UTC
                      Time: 1543575000
                      Value: "0.03755901800347601"
                    - Date: 2018-11-30 10:50:30 +0000 UTC
                      Time: 1543575030
                      Value: "0.03755901800000006"
                    - Date: 2018-11-30 10:51:00 +0000 UTC
                      Time: 1543575060
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:51:30 +0000 UTC
                      Time: 1543575090
                      Value: "0.037559018"
                    - Date: 2018-11-30 10:52:00 +0000 UTC
                      Time: 1543575120
                      Value: "0.037559018"
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
