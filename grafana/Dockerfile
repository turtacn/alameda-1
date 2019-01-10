FROM grafana/grafana:5.4.2

COPY ["dashboards/", "/var/lib/alameda/dashboards"]
COPY ["datasource.yaml", "/etc/grafana/provisioning/datasources/"]
COPY ["dashboard.yaml", "/etc/grafana/provisioning/dashboards/"]
