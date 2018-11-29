# QuickStart

This document guides you from build, deploy to use Alameda.
- [Build Alameda](#build-alameda)
- [Deploy Alameda](#deploy-alameda)
- [Use Alameda](#use-alameda)

# Build Alameda
Please follow the [build](./build.md) guide.

# Deploy Alameda
Please follow the [deploy](./deploy.md) guide.

# Use Alameda

## Specify a target to request Alameda services
User can create a custom resource of *AlamedaResource* custom resource definition (CRD) to instruct Alameda that
1. which Pod(s) to watch by Kubernetes *selector* construct, and
2. what policy that Alameda should use to give recommendations.

Currently Alameda provides *stable* and *compact* policy.
```
apiVersion: v1
kind: Namespace
metadata:
  name: alameda
---
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaResource
metadata:
  name: alameda
  namespace: alameda
spec:
  policy: stable
  enable: true
  selector:
    matchLabels:
      app: nginx
```

You can list all Pods that are watched by Alameda with:
```
$ oc get alamedaresources
```

## Retrieve Alameda prediction and recommendation result
Alameda outputs raw workload metrics prediction and recommendations in a global planning manner for all the pods watched by Alameda.
They are presented as *alamedaresourceprediction* CRD.
You can check Alameda prediction and recommendation results by:
```
$ oc get alamedaresourceprediction
```

## Example

- Deploy a nginx application example by:
```
$ cd <alameda>/example/samples/nginx
$ oc apply -f nginx_deployment.yaml
```
- Request Alameda to predict and recommend the resource usage for nginx Pod by:
```
$ cd <alameda>/example/samples/nginx
$ oc apply -f alameda_deployment.yaml
```
You can check that Alameda is watching the nginx Pod by:
```
$ oc get alamedaresources
NAME      AGE
alameda   5m
$ oc describe alamedaresources alameda
Name:         alameda
Namespace:    alameda
Labels:       <none>
Annotations:  annotation-k8s-controller={
  "DeploymentMap": {
    "82db70d6-f3bb-11e8-8227-52540087d9a4": {
      "UID": "82db70d6-f3bb-11e8-8227-52540087d9a4",
      "Name": "nginx-deployment",
      "PodMap": {
...
              kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"autoscaling.containers.ai/v1alpha1","kind":"AlamedaResource","metadata":{"annotations":{},"name":"alameda","namespace":"alameda"},"spec"...
API Version:  autoscaling.containers.ai/v1alpha1
Kind:         AlamedaResource
Metadata:
  Creation Timestamp:  2018-11-29T09:57:53Z
  Generation:          1
  Resource Version:    23432
  Self Link:           /apis/autoscaling.containers.ai/v1alpha1/namespaces/alameda/alamedaresources/alameda
  UID:                 3ce7cd30-f3bd-11e8-8227-52540087d9a4
Spec:
  Enable:  true
  Policy:  stable
  Selector:
    Match Labels:
      App:  nginx
Status:
Events:  <none>
```
And the prediction and recommendation for this nginx Pod are:
```
$ oc get alamedaresourceprediction
```


