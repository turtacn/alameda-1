# Build Alameda docker images from source code

Running Alameda requires two components:
1. [Alameda operator](https://github.com/containers-ai/alameda) which interacts with Kubernetes cluster
2. [Alameda-ai](https://github.com/containers-ai/alameda-ai) which generates predictions and recommendations with deep learning techniques 

The following steps show how to build Alameda images.
- First we need to install [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and [Docker](https://docs.docker.com/install/#supported-platforms) environment
- Build Alameda operator image and dashboard image by:
    ```
    $ git clone https://github.com/containers-ai/alameda.git
    $ cd alameda/operator
    $ docker build -t operator .
    $ cd ../grafana
    $ docker build -t dashboard .
    ```
- Build Alameda-ai image by:
    ```
    git clone https://github.com/containers-ai/alameda-ai.git
    cd alameda-ai
    docker build -t alameda-ai .
    ```
Then you can find the built *alameda-ai*, *operator*, and *dashboard* images in your docker environment.
```
$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
alameda-ai          latest              3c319e0eed87        7 seconds ago       1.76GB
dashboard           latest              aa3a33126b34        3 minutes ago       244MB
operator            latest              328c486be922        6 minutes ago       44.3MB
<none>              <none>              c47111eaf0a5        7 minutes ago       591MB
python              3.6                 1ec4d11819ad        12 days ago         918MB
golang              1.11.2-alpine       57915f96905a        3 weeks ago         310MB
alpine              latest              196d12cf6ab1        2 months ago        4.41MB
```
