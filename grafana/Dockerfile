FROM grafana/grafana:5.4.3

ARG image_tag

ENV APPHOME=/opt/prophetstor/federatorai/grafana \
    APP_DEBUG=0

COPY . /tmp/src

## avoid define WORKDIR to use original container defination

USER root
RUN set -x \
    && id \
    && apt-get update \
    && apt-get install -y nano curl procps sqlite3 \
    && apt-get autoremove -y && apt-get autoclean -y && rm -rf /var/lib/apt/lists/* 

RUN set -x \
    ## Build
    && (if [ "${image_tag}" = "" ]; then /bin/echo -e "\n\nError. Please run with '--build-arg image_tag=<tag_name>'\n\n"; exit 1; fi) \
    && sed -i -e "s/__DEV_VERSION__/${image_tag}/g" /tmp/src/rootfs/var/lib/grafana/dashboards/Home/home.json \
    ## Install
    && mkdir -p ${APPHOME} \
    && mkdir -p /usr/share/grafana/public/dashboards /var/lib/grafana/dashboards/alameda /var/log/grafana \
    && chmod 777 /usr/share/grafana/public/dashboards /var/lib/grafana/dashboards/alameda \
    && chmod -R 777 /tmp/src/rootfs/var/lib/grafana/ /tmp/src/rootfs/etc/grafana/ \
    && (cd /tmp/src/rootfs; tar czvf ${APPHOME}/alameda-dashboard.tgz .; cd -) \
    && (cd /; tar xzvf ${APPHOME}/alameda-dashboard.tgz; cd -) \
    && ln -sfv /var/lib/grafana/dashboards/Home/home.json /usr/share/grafana/public/dashboards/home.json \
    ## Need permission inside the following directories
    && chmod -R 777 /etc/grafana /usr/share/grafana /var/lib/grafana /var/log/grafana \
    ## OS
    && echo "${USER_NAME}:x:${USER_UID}:0:Federator.ai:${APPHOME}:/bin/sh" >> /etc/passwd \
    ## The following lines startup script will add running user id into /etc/passwd
    && chmod g+w /etc/passwd \
    ## init scripts
    && cp /tmp/src/tini /sbin/tini && chmod 755 /sbin/tini \
    && cp /tmp/src/init.sh /init.sh && chmod 755 /init.sh \
    && rm -rf /tmp/src

USER grafana

ENTRYPOINT ["/sbin/tini","-v", "--"]
CMD ["/init.sh", "start"]
