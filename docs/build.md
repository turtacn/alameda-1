# Build Alameda docker images from source code

## Prerequisites
To build Alameda images from source code, [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [Docker](https://docs.docker.com/install/#supported-platforms) environment are required.
> **Note**: It is recommended to use Docker engine 18.09 or greater

As showed in the [Alameda architecture design](https://github.com/containers-ai/alameda/blob/master/design/architecture.md), Running Alameda requires several components. Some of the components leverage existing open-source solutions such as Prometheus and Grafana. Some of them are implemented in Alameda. The following sections show how to build those components implemented in Alameda from source code.

## Build operator
```
$ git clone https://github.com/containers-ai/alameda.git
$ cd alameda
$ docker build . -t operator:latest -f operator/Dockerfile
```

## Build datahub
```
$ git clone https://github.com/containers-ai/alameda.git
$ cd alameda
$ docker build . -t datahub:latest -f datahub/Dockerfile
```

## Build alameda-ai
```
$ git clone https://github.com/containers-ai/alameda-ai.git
$ cd alameda-ai
$ docker build . -t alameda-ai:latest -f Dockerfile
```

## Check Results
You can find the built *alameda-ai*, *operator*, and *datahub* images in your local docker environment.
```
$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
alameda-ai          latest              2d84db9bf136        6 minutes ago       3.53GB
<none>              <none>              8c52fe903fa2        13 minutes ago      2.63GB
datahub             latest              a3890b881703        26 minutes ago      48.4MB
<none>              <none>              bf7de2af6ce5        26 minutes ago      1.42GB
operator            latest              3999ab5862cd        About an hour ago   48.8MB
<none>              <none>              3de681af2c69        About an hour ago   1.43GB
python              3.6-stretch         70f6aab434cf        45 hours ago        935MB
python              3.6-slim-stretch    eaeeb206b279        10 days ago         138MB
golang              1.11.5-stretch      901414995ecd        3 weeks ago         816MB
alpine              latest              caf27325b298        4 weeks ago         5.53MB
```
