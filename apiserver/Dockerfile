# Build the manager binary
FROM golang:1.11.5-stretch as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda
ADD . .

# Build
RUN cd apiserver && make build

# Copy the controller-manager into a thin image
FROM alpine:latest
ENV AIHOME=/opt/alameda/apiserver
WORKDIR ${AIHOME}
COPY --from=builder /go/src/github.com/containers-ai/alameda/apiserver/install_root.tgz /tmp/
RUN cd / && tar xzvf /tmp/install_root.tgz && rm -fv /tmp/install_root.tgz \
    && chown -R 1001:0 ${AIHOME} \
    && mkdir -p /var/log/alameda && chmod 777 /var/log/alameda

EXPOSE 50055/tcp

ENTRYPOINT ["/opt/alameda/apiserver/bin/apiserver", "run" ]
