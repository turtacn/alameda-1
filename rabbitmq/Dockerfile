FROM openshift/origin-release:golang-1.13 as builder
#RUN yum update -y
# Build the manager binary

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


FROM registry.access.redhat.com/ubi8/ubi
ENV SUMMARY="Alameda RabbitMQ." \
    DESCRIPTION="Alameda RabbitMQ." \
    NAME="Alameda RabbitMQ" \
    VERSION="latest" \
    USER_UID=1001 \
    USER_NAME=alameda

LABEL summary="$SUMMARY" \
      description="$DESCRIPTION" \
      io.k8s.description="$DESCRIPTION" \
      io.k8s.display-name="$NAME" \
      com.redhat.component="customer-container" \
      name="$NAME" \
      version="$VERSION" \
      vendor="ProphetStor Data Services, Inc."


ENV APPHOME="/opt/alameda/alameda-rabbitmq" \
    RABBITMQ_DATA_DIR=/var/lib/rabbitmq \
    RBMQ_MAIN_VER="3.8" \
    RBMQ_DEDUP_PLUGIN_VER="0.4.3" \
    HOME=/var/lib/rabbitmq \
    PATH=/usr/lib/rabbitmq/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    ERLANG_VER="22.2.4"

ENV RBMQ_VER="${RBMQ_MAIN_VER}.2" \
    RBMQ_DEDUP_EZ="rabbitmq_message_deduplication-v3.8.x-${RBMQ_DEDUP_PLUGIN_VER}.ez" \
    RBMQ_ELIXIR_EZ="elixir-1.8.2.ez" \
    ERLANG_RPM="erlang-${ERLANG_VER}-1.el8.x86_64.rpm"

ENV RBMQ_RPM="rabbitmq-server-${RBMQ_VER}-1.el8.noarch.rpm" \
    RBMQ_PLUGIN_FOLDER="/usr/lib/rabbitmq/lib/rabbitmq_server-${RBMQ_VER}/plugins" \
    LANG=C.UTF-8 \
    LANGUAGE=C.UTF-8 \
    LC_ALL=C.UTF-8

COPY rabbitmq/init.sh /init.sh
COPY rabbitmq/prob.sh /prob.sh
COPY --from=builder /go/src/github.com/containers-ai/alameda/rabbitmq/xray.sh /xray.sh
COPY --from=builder /go/src/github.com/containers-ai/alameda/version.txt ${APPHOME}/etc/
COPY --from=builder /go/src/github.com/containers-ai/alameda/rabbitmq/rabbitmqcmd /usr/local/bin/

RUN set -x && yum update -y \
    && yum install -y tar procps-ng psmisc findutils vim-minimal \
    && curl -LO https://github.com/rabbitmq/erlang-rpm/releases/download/v${ERLANG_VER}/${ERLANG_RPM} \
    && yum install ${ERLANG_RPM} -y && rm ${ERLANG_RPM} \
    && yum install -y python2 && ln -sfv /usr/bin/python2 /usr/bin/python \
    && curl -LO https://bintray.com/rabbitmq/rpm/download_file?file_path=rabbitmq-server/\
v${RBMQ_MAIN_VER}.x/el/8/noarch/${RBMQ_RPM} \
    && rpm --import https://www.rabbitmq.com/rabbitmq-release-signing-key.asc \
    && yum install ${RBMQ_RPM} -y && rm ${RBMQ_RPM} \
    && curl -LO https://github.com/noxdafox/rabbitmq-message-deduplication/releases/\
download/${RBMQ_DEDUP_PLUGIN_VER}/${RBMQ_DEDUP_EZ} && mv ${RBMQ_DEDUP_EZ} ${RBMQ_PLUGIN_FOLDER} \
    && curl -LO https://github.com/noxdafox/rabbitmq-message-deduplication/releases/\
download/${RBMQ_DEDUP_PLUGIN_VER}/${RBMQ_ELIXIR_EZ} && mv ${RBMQ_ELIXIR_EZ} ${RBMQ_PLUGIN_FOLDER} \
    && curl -LO https://raw.githubusercontent.com/rabbitmq/rabbitmq-management/v${RBMQ_MAIN_VER}.x/bin/rabbitmqadmin \
    && chmod +x rabbitmqadmin && mv rabbitmqadmin /usr/local/bin/rabbitmqadmin \
    && mkdir -pv /licenses/ && touch /licenses/.placeholder \
    && chmod 755 /init.sh /prob.sh \
    && chmod g+w /etc/passwd /etc \
    && chmod -R 777 "$RABBITMQ_DATA_DIR" /etc/rabbitmq /var/log/rabbitmq \
    && rabbitmq-plugins enable rabbitmq_management rabbitmq_message_deduplication rabbitmq_tracing \
    && rm -fv /.erlang.cookie /var/lib/rabbitmq/.erlang.cookie

USER rabbitmq

CMD ["/init.sh"]
