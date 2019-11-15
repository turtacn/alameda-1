# Build the manager binary
FROM golang:1.13-stretch as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda
ADD . .

# Build datahub executable binary
RUN ["/bin/bash", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'\" -a -o ./datahub/datahub github.com/containers-ai/alameda/datahub/cmd"]

# Build license-utils executable binary
RUN ["/bin/bash", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'\" -a -o ./datahub/license-utils github.com/containers-ai/alameda/datahub/tools/license-utils/cmd"]

# Copy the controller-manager into a thin image
FROM greyfoxit/alpine-glibc

ENV FEDAIHOME="/opt/prophetstor/federatorai" \
    LD_LIBRARY_PATH="/opt/prophetstor/federatorai/lib"

# Copy keycode management tool
RUN mkdir -p /opt/prophetstor/federatorai/bin /opt/prophetstor/federatorai/lib /opt/prophetstor/federatorai/etc /opt/prophetstor/federatorai/var/tmp
WORKDIR /opt/prophetstor/federatorai/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/bin/license_main bin/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/lib/license_helper.so lib/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/lib/license_util.so lib/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/lib/util_crypto.so lib/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/lib/util_ldap.so lib/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes/lib/_utilModule.so lib/

# Copy datahub
WORKDIR /root/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/etc/datahub.toml /etc/alameda/datahub/datahub.toml
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/datahub .

# Copy license-utils
WORKDIR /root/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/license-utils .

EXPOSE 50050/tcp
ENTRYPOINT ["./datahub"]
CMD [ "run" ]
