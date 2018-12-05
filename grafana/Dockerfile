FROM grafana/grafana:5.4.0

COPY ["dashboards/", "/var/lib/alameda/dashboards"]
COPY ["datasource.yaml", "/etc/grafana/provisioning/datasources/"]
COPY ["dashboard.yaml", "/etc/grafana/provisioning/dashboards/"]

# plugin versions: grafana-simple-json-datasource 1.4.0
COPY ["plugins/", "/var/lib/grafana/plugins"]