FROM registry.access.redhat.com/ubi8/ubi

ARG image_tag
ARG GF_VER="6.4.5"
ARG GF_RPM="grafana-${GF_VER}-1.x86_64.rpm"

ENV APPHOME=/opt/prophetstor/federatorai/grafana \
    APP_DEBUG=0 \
    GF_PANELS_DISABLE_SANITIZE_HTML=true \
    PATH="/usr/share/grafana/bin:$PATH" \
    GF_PATHS_CONFIG="/etc/grafana/grafana.ini" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning"

COPY . /tmp/src

## avoid define WORKDIR to use original container defination

RUN set -x \
    && id \
    && yum update -y \
    && curl -LO https://dl.grafana.com/oss/release/$GF_RPM \
    && yum localinstall $GF_RPM -y && rm $GF_RPM \
    && curl -LO https://raw.githubusercontent.com/grafana/grafana/v${GF_VER}/packaging/docker/run.sh \
    && chmod +x run.sh \
    && yum install nano curl procps sqlite nginx -y \
    && yum clean all && rm -rf /var/cache/yum \
    && rm -fv /etc/nginx/conf.d/* \
    && mv -fv /tmp/src/nginx.conf /etc/nginx/nginx.conf \
    && mv -fv /tmp/src/ssl.crt /etc/nginx/ssl.crt \
    && mv -fv /tmp/src/ssl.key /etc/nginx/ssl.key \
    ## Need write permission on the following directories
    && mkdir -p /var/tmp/nginx /var/lib/nginx && chmod -R 777 /var/log /var/tmp/nginx /var/lib/nginx /etc/nginx \
    && mkdir -p "$GF_PATHS_PLUGINS" \
    && chown -R grafana:grafana "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" \
    && chmod -R 777 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" \
    && chmod 777 "$GF_PATHS_CONFIG"

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
    ## OS
    && echo "${USER_NAME}:x:${USER_UID}:0:Federator.ai:${APPHOME}:/bin/sh" >> /etc/passwd \
    ## The following lines startup script will add running user id into /etc/passwd
    && chmod g+w /etc/passwd \
    ## init scripts
    && cp /tmp/src/xray.sh /xray.sh \
    && cp /tmp/src/tini /sbin/tini && chmod 755 /sbin/tini \
    && cp /tmp/src/init.sh /init.sh && chmod 755 /init.sh \
    && rm -rf /tmp/src

USER grafana

ENTRYPOINT ["/sbin/tini","-v", "--"]
CMD ["/init.sh", "start"]
