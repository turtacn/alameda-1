# Architecture

Alameda Operator leverages kubebuilder to create controller to reconcile kubernetes resources. Currently, there are 6 types of controller instances running in the code, reside under [folder](./pkg/controller). These controller instances are listed below.
*   AlamedaRecommendation
*   AlamedaScaler
*   Deployment
*   DeploymentConfig
*   Node
*   StatefulSet

Alameda Operator lists **Pods** owned by each k8s/openshift **Workload Controller** (Deployment, DeploymentConfig and StatefulSet) in **AlamedaScaler.Status**. To simplify and centralize this process, Alameda Operator only updates **AlamedaScaler.Status** in AlamedaScaler controller instance, other controller instances (Deployment, DeploymentConfig and StatefulSet) just try to trigger the process by updating the **AlamedaScaler.Spec.CustomResourceVersion**.

## AlamedaRecommendation

The alamedaRecommendation controller watches AlamedaRecommendation type. It deletes the AlamedaRecommendation which has no mapping Pod in the AlamedaScaler. If the mapping Pod is found, fetches the latest resources recommendation of the Pod (samne namespace and name with the AlamedaRecommendation) from Alameda Datahub and stores the data in **AlamedaRecommendation.Spec.Containers**.

## AlamedaScaler

The alamedaScaler controller watches AlamedaScaler type, it refreshs AlamedaScaler's status and update the dependent resources in the Reconcile function. Currently, the dependent resources are **AlamedaRecommendation** that maps to each Pod in the status, and the registration data that tells **Alameda Datahub** which **Pod** and **Workload Controller** that needs to be predicted. 

## Deployment

The deployment controller watches Deployment type, it tries to update the CustomResourceVersion in the spec of the AlamedaScaler which is found monitoring the Deployment.

## DeploymentConfig

The deploymentConfig controller watches DeploymentConfig type, it tries to update the CustomResourceVersion in the spec of the AlamedaScaler which is found monitoring the DeploymentConfig.

## Node 

The node controller watches Node type. It provides the detail information of the node to **Alameda Datahub**.

## StatefulSet

The statefulSet controller watches StatefulSet type, it tries to update the CustomResourceVersion in the spec of the AlamedaScaler which is found monitoring the StatefulSet.