#!/bin/bash

for pt in `find . | grep \\\.proto$ | grep -v ^\\\./include`; \
        do
            python -m grpc_tools.protoc -I . -I include/ --python_out=./ --grpc_python_out=./ $pt;
            protoc -I . -I include/ $pt --go_out=plugins=grpc:.;
        done

