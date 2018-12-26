FROM influxdb:1.7.2-alpine

ENV INFLUXDB_ADMIN_ENABLED=true INFLUXDB_ADMIN_USER=alameda INFLUXDB_ADMIN_PASSWORD=alameda INFLUXDB_USER=alameda INFLUXDB_USER_PASSWORD=alameda

COPY ["init/", "/docker-entrypoint-initdb.d"]