# api

Alameda API definitions of Alameda-ai service and Alameda operator

## Prerequisition

1. Install [protoc](https://github.com/protocolbuffers/protobuf/releases) command for generating golang code
2. Install the packages for generating python code
    ```bash
    pip install -r requirements.txt
    ```


## How to compile

1. Run the script to generate client and service code for golang and python
    ```bash
    ./compile_proto.sh
    ```
The generated code will be located in the same folder as the .proto files.

## How to use

### Coding with golang

Add the following import declarations in your .go files when using the Alameda API gRPC calls.
```
import "github.com/containers-ai/api/alameda_api/v1alpha1/ai_service"
import "github.com/containers-ai/api/alameda_api/v1alpha1/operator"
```

## Coding with python

Install alameda-api packages by
```bash
pip install git+https://github.com/containers-ai/api.git
```
Then you can use Alameda API gRPC calls in your .py files by
```
from alameda_api.v1alpha1.ai_service import ai_service_pb2, ai_service_pb2_grpc
from alameda_api.v1alpha1.operator import server_pb2, server_pb2_grpc

```
