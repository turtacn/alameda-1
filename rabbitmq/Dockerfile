# Build the manager binary
FROM openshift/origin-release:golang-1.13 as builder
RUN yum update -y

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/containers-ai/alameda

COPY . .

# Build
RUN ["/bin/bash", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'\" -a -o ./rabbitmq/rabbitmqcmd github.com/containers-ai/alameda/rabbitmq/cmd"]


FROM docker.io/rabbitmq:3.8.2-management-alpine

COPY rabbitmq/plugins/ plugins/
COPY rabbitmq/init.sh /init.sh
COPY rabbitmq/prob.sh /prob.sh
COPY --from=builder /go/src/github.com/containers-ai/alameda/rabbitmq/rabbitmqcmd /usr/local/bin/
RUN chmod 755 /init.sh /prob.sh && \
    rabbitmq-plugins enable rabbitmq_message_deduplication rabbitmq_tracing

CMD ["/init.sh"]

