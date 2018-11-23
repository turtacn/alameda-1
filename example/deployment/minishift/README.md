# Use Minishit for local development environment

---

## Pre-requirement

#### Hypervisor Choice

**Mac**: xhyve

**linux**: KVM

**windows**: hyper-v if using windows pro

**all platform support**:virtualbox

#### Setup virtual environment
Setup for each platform details please reference openshit document

https://docs.okd.io/latest/minishift/getting-started/setting-up-virtualization-environment.html

#### Install minishift

Install minishift please reference openshift document for each platform:

https://docs.okd.io/latest/minishift/getting-started/installing.html

## Setup Developer Environment

Execute below command to start minishift
```
minishift start
```

After install execute below command to use oc command
```
eval $(minishift oc-env)
```
To access web console
```
minishift console
```

## Setup minishift addons

Enable admin addons
```
minishift addons enable admin-user
minishift addons apply admin-user
```

After this login as admin, default password: **password**

```
oc login -u admin
```

Enable registry-route addons

```
minishift addons enable registry-route
minishift addons apply registry-route
```

Minishift registry using self-signed CA certification,
therefore we need add insecure registry for docker login

First execute command
```
minishift openshift registry
```
copy paste address for setting insecure-registries

**Mac**:

Add in docker preference daemon tab

![mac](img/mac.png "mac")

**Windows**

Add in docker preference daemon tab

![windows](img/windows.png "windows")

**linux**

add "insecure-registries" : ["<address\>"] to /etc/docker/daemon.json
than restart docker daemon

## Setup project

Login as developer to create project
```
oc login -u developer
```
Create alameda project
```
oc new-project alameda
```

## Apply yaml for kubernetes config

Login as admin, then apply follow yaml files in example/deployment/minishift

```
oc login -u admin
```
```
oc adm policy add-scc-to-user anyuid system:serviceaccount:opsmx:tiller
oc adm policy add-scc-to-group anyuid system:authenticated
oc apply -f prometheus.yaml
oc apply -f rbac
oc apply -f crds
oc apply -f service
oc apply -f deployconfig
```

Clone code and build docker image for operator
```
git clone https://github.com/containers-ai/alameda.git
cd alameda/operator
docker build -t operator .
```
Clone code and build docker image for alameda-ai
```
git clone https://github.com/prophetstor-ai/alameda-ai.git
cd alameda-ai
docker build -t alameda-ai .
```
Tag image and push to minishift registry
```
oc login -u developer
docker tag operator $(minishift openshift registry)/alameda/operator
docker tag alameda-ai $(minishift openshift registry)/alameda/alameda-ai
oc whoami -t | docker login -u developer --password-stdin $(minishift openshift registry)
docker push $(minishift openshift registry)/alameda/operator
docker push $(minishift openshift registry)/alameda/alameda-ai
```

Check pod successful created and running
```
oc get pod
```
log in to pod
```
oc exec -it <pod name> bash
```

After ensure both operator and alameda-ai running execute below command
```
oc apply -f samples/nginx_deployment.yaml
oc apply -f samples/alameda_deployment.yaml
```
