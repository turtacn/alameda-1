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
