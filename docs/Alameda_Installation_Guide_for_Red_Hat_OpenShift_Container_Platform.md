<P style="font-size:24pt; bold">Alameda Installation Guide for Red Hat OpenShift Container Platform</p>

# Prerequisites
1.**Platform Requirement**

OpenShift

2.**OpenShift Cluster Admin User**

A user bound with the "cluster-admin" role is needed for deployment (no longer needed afterward)
```
$ oc adm policy add-cluster-role-to-user cluster-admin <user_name>
```

3.**OpenShift Persistent Volumes**

Cluster admin needs to prepare **3 Persistent Volumes (PV)** for Alameda OpenShift template to do **Persistent Volume Claims(PVC)**. PV needs to meet following requirements:

* **InfluxDB PV**:
	1. Access Mode: ReadWriteOnce
	2. Capacity: Meets requirement for PVC capacity inputed during deployment (at least 10GB)
* **Alameda-AI PV**:
	1. Access Mode: ReadWriteOnce
	2. Capacity: Meets requirement for PVC capacity inputed during deployment (at least 10GB)
* **Grafana PV**:
	1. Access Mode: ReadWriteOnce
	2. Capacity: Meets requirement for PVC capacity inputed during deployment (at least 2GB)

Note: Check **Alameda Installation step 7** for detail about specifying PVCs capacity to meet the PVs created here.
# Alameda Installation

1. Go to OpenShift GUI and login using a user with the **cluster-admin** role binding.

2. Create a project. (ex: alameda)
	
	![](./img/openshift_guide/1.png)
	
3. Select the newly created project and choose **Import YAML/JSON**.
	
	![](./img/openshift_guide/2.png)
	
4. Download alameda.yml from github [openshift/template/deploy/](../openshift/template/deploy/) and either upload the file or just directly paste the content into the text area.

	![](./img/openshift_guide/3.png)
	
5. De-select **Process the template**, select **Save template**.
	
	![](./img/openshift_guide/4.png)
	
6. After you've finished importing, go to ***Catalog*** in the sidebar, and you should see the **Federator.ai** icon.
	
	![](./img/openshift_guide/5.png)
	
7. Click on the Federator.ai icon then click Next. Fill in the project where alameda will be deployed as **alameda namespace** (ex. alameda) - note this needs to match the current project being accessed. For the **DockerHub config json** secret key and **Alameda image tag**, please obtain them from a ProphetStor sales representative. You can also specify the InfluxDB database capacity, Alameda-AI capacity, and Grafana database capacity if you want but remember to match the PVs size which you created in **Prerequisites** section.
	
	![](./img/openshift_guide/6.png)
	
8. Click **Continue Anyway** to ignore the warning.
	
	![](./img/openshift_guide/7.png)
	
9. Go to ***Overview*** to see the deployment take effect.
	
	![](./img/openshift_guide/8.png)
	
10. After deployment is done, switch to the openshift-monitoring project (a default project in OCP). Go to **Resources > Secrets > grafana-datasources** and reveal the secret.
	
	![](./img/openshift_guide/9.png)
	
11. Write down the following three values from prometheus.yaml.

	| Name | Value (example)|
	| --- | --- |
	| url | https://prometheus-k8s.openshift-monitoring.svc:9091 |
	| basicAuthUser | internal |
	| basicAuthPassword | ******* |
	
	![](./img/openshift_guide/10.png)
	
12. Go back to the alameda project to locate the URL of the alameda-grafana pod. Open that address in a browser. (default username/password is admin/admin).
	
	![](./img/openshift_guide/11.png)
	
	![](./img/openshift_guide/12.png)
	
13. Go to **Configuration > Data Sources** and click on the **Add data source** button. Select Promethues and fill in the previously recorded url, select **Basic Auth** and fill in AuthUser and AuthPassword. Also, remember to check **Skip TLS verify**. Click the **Save & Test** button.
	
	![](./img/openshift_guide/13.png)
	
	![](./img/openshift_guide/14.png)
	
14. Go to the plus button to import the grafana dashboard json files in [helm/grafana/dashboards](../helm/grafana/dashboards/)

	![](./img/openshift_guide/15.png)

# Alameda Configuration

1. Go to the project where you would like resource predictions and recommendations. Go to its **Deployment or DeploymentConfig** and edit the yaml file. Record the labels value of Deployment or DeploymentConfig (ex: **app: ocp-smoke-test**) 
	
	![](./img/openshift_guide/16.png)
	
	![](./img/openshift_guide/17.png)
	
2. Click on **Add to Project** and select **Import YAML/JSON**.
	
	![](./img/openshift_guide/18.png)
	
3. Put in the following yaml information to tell **AlamedaScaler** which Deployment/DeploymentConfig is applicable and click the **Create** button. (Click **Continue Anyway** to ignore the warning.)

```
apiVersion: autoscaling.containers.ai/v1alpha1
kind: AlamedaScaler
metadata:
    name: alameda
    namespace: ocp-smoke-test #(YOUR_PROJECT_NAME)
spec:
    policy: stable
    enable: true
    selector:
        matchLabels:
            app: ocp-smoke-test #(Your Deployment/DeploymentConfig labels)
```
![](./img/openshift_guide/19.png)

4. Use the following oc command to check if **AlamedaScaler** successfully found the pods under the desired project. You will see **deploymentconfigs:{}** and **deployments:{}** if the AlamedaScaler didn't find any applicable pods.
	
	![](./img/openshift_guide/20.png)
	
