# Build Alameda docker images from source code

## Prerequisites
To build Alameda images from source code, [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [Docker](https://docs.docker.com/install/#supported-platforms) environment are required.
> **Note**: It is recommended to use Docker engine 18.09 or greater

As showed in the [Alameda architecture design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), Running Alameda requires several components. The following sections show how to build them from source code.

## Build operator
```
$ git clone https://github.com/containers-ai/alameda.git
$ cd alameda/operator
$ docker build -t operator .
```

## Build alameda-ai
```
git clone https://github.com/containers-ai/alameda-ai.git
cd alameda-ai
docker build -t alameda-ai .
```

## Build Grafana

Alameda adds templates and json backend customization to visualize predicted metrics. Please use the following commands to build a customized Grafana image.
```
git clone https://github.com/containers-ai/alameda.git
cd grafana
docker build -t grafana .
```

## Check Results
You can find the built *alameda-ai*, *operator* and *Grafana* images in your local docker environment.
```
$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
alameda-ai          latest              3c319e0eed87        7 seconds ago       1.76GB
operator            latest              328c486be922        6 minutes ago       44.3MB
grafana             latest              f80f990fa61c        10 days ago         244MB
<none>              <none>              c47111eaf0a5        7 minutes ago       591MB
python              3.6                 1ec4d11819ad        12 days ago         918MB
golang              1.11.4-alpine       57915f96905a        3 weeks ago         310MB
alpine              latest              196d12cf6ab1        2 months ago        4.41MB
```
