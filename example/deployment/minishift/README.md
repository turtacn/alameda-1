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

 add in docker preference daemon tab

**Windows**

windows add in docker preference daemon tab

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

Login as admin, then apply follow yaml files in config/minishift

```
oc apply -f prometheus.yaml
oc apply -f rbac
oc apply -f crds
oc apply -f manager
oc apply -f deployconfig
```

Clone code and build docker image
```
git clone https://github.com/containers-ai/alameda.git
cd alameda/operator
docker build -t operator .
https://github.com/prophetstor-ai/alameda-ai.git
cd alameda-ai
docker build -t alameda-ai .
```
Tag image and push to minishift registry
```
docker tag operator <address>/alameda/operator
docker tag alameda-ai as <address>/alameda/alameda-ai
docker push <address>/alameda/operator
docker push <address>/alameda/alameda-ai
```
