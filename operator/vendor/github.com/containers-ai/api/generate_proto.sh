#!/bin/bash

for pt in `find . | grep \\\.proto$ | grep -v ^\\\./include`; \
        do
            python -m grpc_tools.protoc -I . -I include/ --python_out=python/containers-ai_api --grpc_python_out=python/containers-ai_api $pt;
        done

for pt in `find . | grep \\\.proto$ | grep -v ^\\\./include`; do protoc -I . -I include/ $pt --go_out=plugins=grpc:.; done
