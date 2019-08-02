FROM grafana/grafana:5.4.3

ENV APPHOME=/opt/prophetstor/federatorai/grafana \
    APP_DEBUG=0

WORKDIR ${APPHOME}

USER root
RUN set -x \
    && id \
    && apt-get update \
    && apt-get install -y nano curl procps sqlite3 \
    && apt-get autoremove -y && apt-get autoclean -y && rm -rf /var/lib/apt/lists/* 

COPY ./rootfs/usr/share/grafana/public/dashboards/home.json /usr/share/grafana/public/dashboards/home.json

RUN set -x \
    && chmod 777 /usr/share/grafana/public/dashboards/* \
    && touch /var/lib/grafana/grafana.db \
    && chmod -R 777 /var/lib/grafana/

USER grafana
