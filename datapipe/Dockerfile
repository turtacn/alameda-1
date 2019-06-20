# Build the manager binary
FROM golang:1.11.5-stretch as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda
ADD . .

# Build
RUN cd datapipe && make build

# Prepare the package into a thin image
FROM alpine:latest
ENV AIHOME=/opt/alameda/datapipe \
    USER_UID=1001 \
    USER_NAME=alameda

WORKDIR ${AIHOME}
COPY --from=builder /go/src/github.com/containers-ai/alameda/datapipe/install_root.tgz /tmp/

RUN set -x \
    && apk add --no-cache tini logrotate && rm -rfv /var/cache/apk/* \
    && adduser -h ${AIHOME} -g "Federator.ai" -s /bin/sh -G root -D -u ${USER_UID} ${USER_NAME} \
    # The following lines for logrotate - startup script will add running user id into /etc/passwd
    && chmod g+w /etc/passwd \
    # logrotate need writable permission on "/var/lib/"
    && chmod ug+w /var/lib /var/log \
    # install packages
    && cd / && tar xzvf /tmp/install_root.tgz && rm -fv /tmp/install_root.tgz \
    && chown -R ${USER_UID}:root ${AIHOME} && chmod -R ug+w ${AIHOME} \
    && mkdir -pv -m 775 /var/log/alameda && chown -R ${USER_UID}:root /var/log/alameda

EXPOSE 50060/tcp

USER ${USER_UID}
ENTRYPOINT ["/sbin/tini","-v", "--"]
CMD ["/init.sh"]
