# Build the manager binary
FROM openshift/origin-release:golang-1.13 as builder
#RUN yum update -y

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/containers-ai/alameda

COPY . .
RUN set -x \
    && export GIT_COMMIT=`git rev-parse --short HEAD || echo 'none'` \
    && export _TMP_STR=`git show-ref --abbrev --head | grep "^${GIT_COMMIT} " | grep -v "HEAD$" | head -1 | awk '{print $NF}'` \
    && [ "${_TMP_STR}" = "" ] && export GIT_REF="HEAD" || export GIT_REF="${_TMP_STR}" \
    && [ "`git diff --shortstat`" = "" ] && export GIT_DIRTY="" || export GIT_DIRTY="-dirty" \
    && export CODE_VERSION="${GIT_REF}@${GIT_COMMIT}${GIT_DIRTY}" \
    && echo "CODE_VERSION=${CODE_VERSION}" > version.txt

# Build
RUN ["/bin/bash", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'\" -a -o ./rabbitmq/rabbitmqcmd github.com/containers-ai/alameda/rabbitmq/cmd"]


FROM docker.io/rabbitmq:3.8.2-management-alpine

ENV AIHOME="/opt/alameda/alameda-datahub"

COPY rabbitmq/plugins/ plugins/
COPY rabbitmq/init.sh /init.sh
COPY rabbitmq/prob.sh /prob.sh
COPY --from=builder /go/src/github.com/containers-ai/alameda/rabbitmq/xray.sh /xray.sh
COPY --from=builder /go/src/github.com/containers-ai/alameda/version.txt ${AIHOME}/etc/
COPY --from=builder /go/src/github.com/containers-ai/alameda/rabbitmq/rabbitmqcmd /usr/local/bin/
RUN chmod 755 /init.sh /prob.sh && \
    rabbitmq-plugins enable rabbitmq_message_deduplication rabbitmq_tracing \
    && apk update && apk add --no-cache curl && rm -rf /var/cache/apk/*

CMD ["/init.sh"]

