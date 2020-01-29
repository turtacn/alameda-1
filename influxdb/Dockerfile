FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV AIHOME=/opt/alameda/alameda-influxdb \
    INFLUXDB_HTTP_HTTPS_ENABLED=true \
    INFLUXDB_HTTP_HTTPS_CERTIFICATE=/etc/ssl/certificate.pem \
    INFLUXDB_HTTP_HTTPS_PRIVATE_KEY=/etc/ssl/certificate.key \
    INFLUXDB_ADMIN_ENABLED=true \
    INFLUXDB_ADMIN_USER=alameda \
    INFLUXDB_ADMIN_PASSWORD=alameda \
    INFLUXDB_USER=alameda \
    INFLUXDB_USER_PASSWORD=alameda

COPY xray.sh /xray.sh
COPY influxdb.repo /etc/yum.repos.d/influxdb.repo
COPY ["init/", "/docker-entrypoint-initdb.d"]
COPY entrypoint.sh /entrypoint.sh
COPY init-influxdb.sh /init-influxdb.sh

RUN set -x \
    && chmod +x /entrypoint.sh /init-influxdb.sh \
    && microdnf update -y \
    && microdnf install -y vim openssl influxdb \
    && microdnf clean all \
    && openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout $INFLUXDB_HTTP_HTTPS_PRIVATE_KEY -out $INFLUXDB_HTTP_HTTPS_CERTIFICATE \
      -subj "/C=  /ST= /L= /O= /OU= /CN= /emailAddress= "

EXPOSE 8086/TCP

ENTRYPOINT ["/entrypoint.sh"]
CMD ["influxd"]
