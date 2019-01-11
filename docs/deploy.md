# Deployment

This deploy guide provides step-by-step instructions for
- Kubernetes  
- OpenShift Origin  
- Minishift

And you can also choose to deploy Federator.ai with Helm charts.

## with Helm charts

It just needs a few seconds to deploy Federator.ai with Helm charts. Please refer to the [README](./helm/README.md) for more details.

> **Note**: Helm chart deployment is applicable to any Kubernetes distributions including openshift.

## on Kubernetes
Please refer to the [README](../example/deployment/kubernetes/README.md).

## on OpenShift Origin

This section shows how to deploy Federator.ai from source code to a single-node OpenShift Origin cluster.

- Build Federator.ai docker images by following the [build](./build.md) guide. Supposely you will have *operator*, *datahub* and *alameda-ai* images in your docker environment as:
    ```
    $ docker images
    REPOSITORY                    TAG                 IMAGE ID            CREATED             SIZE
    dashboard                     latest              aa3a33126b34        20 hours ago        244MB
    operator                      latest              58e9fad95e3d        21 hours ago        44.3MB
    datahub                       latest              bcabb1da9ed8        10 minutes ago      41.4MB
    alameda-ai                    latest              a366398fa9fc        45 hours ago        1.76GB
    ```
- Prepare a running OpenShift cluster. Please refer to [OpenShift Origin installation guide](https://docs.okd.io/latest/getting_started/administrators.html) for more details. The following steps illustrate creating an OpenShift origin single-node cluster on ubuntu 18.04.
    - Install [Docker](https://docs.docker.com/install/#supported-platforms) 
    - Add ```{"insecure-registries": ["172.30.0.0/16"]}``` to */etc/docker/daemon.json* and restart docker
    - Download [Openshift Origin v3.11.0](https://github.com/openshift/origin/releases/tag/v3.11.0), untar it and add the extracted directory to system PATH
    - Execute
        ```
        $ oc cluster up
        ```
        Then an openshift single-node cluster should be up and running. Check if all Pods are running without error by:
        ```
        $ oc login -u system:admin
        $ oc get pods --all-namespaces
        NAMESPACE                       NAME                                                      READY     STATUS      RESTARTS   AGE
        default                         docker-registry-1-deploy                                  1/1       Running     0          2m
        default                         persistent-volume-setup-ldhmq                             0/1       Completed   0          3m
        default                         router-1-deploy                                           1/1       Running     0          2m
        kube-dns                        kube-dns-ldwzm                                            1/1       Running     0          4m
        kube-proxy                      kube-proxy-f44qr                                          1/1       Running     0          4m
        kube-system                     kube-controller-manager-localhost                         1/1       Running     0          4m
        kube-system                     kube-scheduler-localhost                                  1/1       Running     0          4m
        kube-system                     master-api-localhost                                      1/1       Running     0          4m
        kube-system                     master-etcd-localhost                                     1/1       Running     0          4m
        openshift-apiserver             openshift-apiserver-6zjlv                                 1/1       Running     0          4m
        openshift-controller-manager    openshift-controller-manager-zc5ws                        1/1       Running     0          3m
        openshift-core-operators        openshift-service-cert-signer-operator-6d477f986b-84wgp   1/1       Running     0          4m
        openshift-core-operators        openshift-web-console-operator-57986c9c4f-7pszp           1/1       Running     0          3m
        openshift-service-cert-signer   apiservice-cabundle-injector-8ffbbb6dc-7cls6              1/1       Running     0          4m
        openshift-service-cert-signer   service-serving-cert-signer-668c45d5f-dkjs9               1/1       Running     0          4m
        openshift-web-console           webconsole-7b454b4b5-7sssd                                1/1       Running     0          2m
        ```
        Note: If *docker-registry* or *router* Pods are not running, try to restart them by:
        ```
        $ oc login -u system:admin
        $ oc project default
        $ oc rollout latest dc/docker-registry
        $ oc rollout latest dc/router
        ``` 
- Login as *system:admin* and create *admin* user with cluster-admin role
    ```
    $ oc login -u system:admin
    $ oc create user admin --full-name=admin
    $ oc adm policy add-cluster-role-to-user cluster-admin admin
    ```
- Create *alameda* project with *admin* user and push Federator.ai images to OpenShift integrated registry
    ```
    $ oc login -u admin
    $ oc new-project alameda
    $ docker tag operator 172.30.1.1:5000/alameda/operator
    $ docker tag datahub 172.30.1.1:5000/alameda/datahub
    $ docker tag alameda-ai 172.30.1.1:5000/alameda/alameda-ai
    $ docker tag dashboard 172.30.1.1:5000/alameda/dashboard
    $ docker login -u admin -p `oc whoami -t` 172.30.1.1:5000
    $ docker push 172.30.1.1:5000/alameda/operator
    $ docker push 172.30.1.1:5000/alameda/datahub
    $ docker push 172.30.1.1:5000/alameda/alameda-ai
    $ docker push 172.30.1.1:5000/alameda/dashboard
    ```
    Check if the imagestreams are creted in alameda namespace by:
    ```
    $ oc get is
    NAME         DOCKER REPO                          TAGS      UPDATED
    alameda-ai   172.30.1.1:5000/alameda/alameda-ai   latest    8 seconds ago
    dashboard    172.30.1.1:5000/alameda/dashboard    latest    3 minutes ago
    operator     172.30.1.1:5000/alameda/operator     latest    3 minutes ago
    datahub      172.30.1.1:5000/alameda/datahub      latest    3 minutes ago
    ```
- Deploy Prometheus by:
    ```
    $ cd <alameda>/example/deployment/openshift
    $ oc apply -f prometheus.yaml
    ```
- Deploy Federator.ai by:
    ```
    $ cd <alameda>/example/deployment/openshift
    $ oc adm policy add-scc-to-user anyuid system:serviceaccount:opsmx:tiller
    $ oc adm policy add-scc-to-group anyuid system:authenticated
    $ oc apply -f rbac
    $ oc apply -f crds
    $ oc apply -f service
    $ oc apply -f deployconfig
    ```
    Check if Federator.ai *operator*, *datahub* and *alameda-ai* Pods are running or not.
    ```
    $ oc get pods -n alameda
    NAME                 READY     STATUS    RESTARTS   AGE
    alameda-ai-1-smnmk   1/1       Running    0         49s
    dashboard-1-vshlj    1/1       Running    0         3m
    operator-1-fg9gx     1/1       Running    0         4m
    datahub-1-tc9he      1/1       Running    0         4m
    ```
## on Minishift

This section shows how to deploy Federator.ai from source code to a Minishift environment.

- Build Federator.ai docker images by following the [build](./build.md) guide. Supposely you will have *operator*, *datahub* and *alameda-ai* images in your docker environment as:
    ```
    $ docker images
    REPOSITORY                    TAG                 IMAGE ID            CREATED             SIZE
    dashboard                     latest              aa3a33126b34        20 hours ago        244MB
    operator                      latest              58e9fad95e3d        21 hours ago        44.3MB
    datahub                       latest              bcabb1da9ed8        10 minutes ago      41.4MB
    alameda-ai                    latest              a366398fa9fc        45 hours ago        1.76GB
    ```
    Export the built Federator.ai images for later use:
    ```
    $ docker save -o alameda-ai.tar alameda-ai:latest
    $ docker save -o operator.tar operator:latest
    $ docker save -o datahub.tar datahub:latest
    $ docker save -o dashboard.tar dashboard:latest
    ```
- Prepare a running Minishift environment. Please refer to the [Minishift Installation guide](https://docs.okd.io/latest/minishift/getting-started/installing.html) for more details. The following steps illustrate creating a Minishift v1.27.0 environment on ubuntu 18.04
    - Download [Minishift v1.27.0](https://github.com/minishift/minishift/releases/download/v1.27.0/minishift-1.27.0-linux-amd64.tgz), untar it and add the extracted directory to system PATH
    - Execute
        ```
        $ minishift start
        ```
        After the command returns, you should have an up and running Minishift environment. Check if all Pods are running without error by:
        ```
        $ eval $(minishift oc-env)
        $ oc login -u system:admin
        $ oc get pods --all-namespaces
        NAMESPACE                       NAME                                                      READY     STATUS      RESTARTS   AGE
        default                         docker-registry-1-deploy                                  1/1       Running     0          2m
        default                         persistent-volume-setup-ldhmq                             0/1       Completed   0          3m
        default                         router-1-deploy                                           1/1       Running     0          2m
        kube-dns                        kube-dns-ldwzm                                            1/1       Running     0          4m
        kube-proxy                      kube-proxy-f44qr                                          1/1       Running     0          4m
        kube-system                     kube-controller-manager-localhost                         1/1       Running     0          4m
        kube-system                     kube-scheduler-localhost                                  1/1       Running     0          4m
        kube-system                     master-api-localhost                                      1/1       Running     0          4m
        kube-system                     master-etcd-localhost                                     1/1       Running     0          4m
        openshift-apiserver             openshift-apiserver-6zjlv                                 1/1       Running     0          4m
        openshift-controller-manager    openshift-controller-manager-zc5ws                        1/1       Running     0          3m
        openshift-core-operators        openshift-service-cert-signer-operator-6d477f986b-84wgp   1/1       Running     0          4m
        openshift-core-operators        openshift-web-console-operator-57986c9c4f-7pszp           1/1       Running     0          3m
        openshift-service-cert-signer   apiservice-cabundle-injector-8ffbbb6dc-7cls6              1/1       Running     0          4m
        openshift-service-cert-signer   service-serving-cert-signer-668c45d5f-dkjs9               1/1       Running     0          4m
        openshift-web-console           webconsole-7b454b4b5-7sssd                                1/1       Running     0          2m
        ```
        Note: If *docker-registry* or *router* Pods are not running, try to restart them by:
        ```
        $ eval $(minishift oc-env)
        $ oc login -u system:admin
        $ oc project default
        $ oc rollout retry dc/docker-registry
        $ oc rollout retry dc/router
        ```
- Add an *admin* user with cluster-admin role
    ```
    $ minishift addons apply admin-user
    ```
- Create *alameda* project with *admin* user and push Federator.ai images to OpenShift integrated registry
    ```
    $ eval $(minishift docker-env)
    $ docker load -i operator.tar
    $ docker load -i datahub.tar
    $ docker load -i alameda-ai.tar
    $ docker tag operator $(minishift openshift registry)/alameda/operator
    $ docker tag operator $(minishift openshift registry)/alameda/datahub
    $ docker tag alameda-ai $(minishift openshift registry)/alameda/alameda-ai
    $ docker tag dashboard $(minishift openshift registry)/alameda/dashboard
    $ oc login -u admin
    $ oc new-project alameda
    $ docker login -u admin -p `oc whoami -t` $(minishift openshift registry)
    $ docker push $(minishift openshift registry)/alameda/operator
    $ docker push $(minishift openshift registry)/alameda/datahub
    $ docker push $(minishift openshift registry)/alameda/alameda-ai
    $ docker push $(minishift openshift registry)/alameda/dashboard
    ```
    Check if the imagestreams are creted in alameda namespace by:
    ```
    $ oc get is
    NAME         DOCKER REPO                          TAGS      UPDATED
    alameda-ai   172.30.1.1:5000/alameda/alameda-ai   latest    2 minutes ago
    dashboard    172.30.1.1:5000/alameda/dashboard    latest    3 minutes ago
    operator     172.30.1.1:5000/alameda/operator     latest    4 minutes ago
    datahub      172.30.1.1:5000/alameda/datahub      latest    4 minutes ago
    ```
- Deploy Prometheus by:
    ```
    $ cd <alameda>/example/deployment/openshift
    $ oc apply -f prometheus.yaml
    ```
- Deploy Federator.ai by:
    ```
    $ cd <alameda>/example/deployment/openshift
    $ oc adm policy add-scc-to-user anyuid system:serviceaccount:opsmx:tiller
    $ oc adm policy add-scc-to-group anyuid system:authenticated
    $ oc apply -f rbac
    $ oc apply -f crds
    $ oc apply -f service
    $ oc apply -f deployconfig
    ```
    Check if Federator.ai *operator*, *datahub* and *alameda-ai* Pods are running or not.
    ```
    $ oc get pods -n alameda
    NAME                 READY     STATUS    RESTARTS   AGE
    alameda-ai-1-7ktxz   1/1       Running   0          1m
    dashboard-1-vshlj    1/1       Running   0          1m
    operator-1-9tlgr     1/1       Running   0          1m
    datahub-1-tc9he      1/1       Running   0          4m
    ```
