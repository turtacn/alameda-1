#!/usr/bin/env bash

# requires helm to be installed

if [[ ${#@} < 2 ]]; then
    echo "Usage: $0 chart values"
#    echo "* semver: semver-formatted version for this package"
    echo "* chart: the directory to output the chart"
    echo "* values: the values file"
    exit 1
fi

#version=$1
chartdir=$1
values=$2

NAMESPACE=alameda
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'charttmpdir'`
tmpchartdir=${tmpdir}
tmpoutputdir=${tmpdir}/output
namespaceResource="
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
"
#cp -R helm/alameda ${tmpchartdir}
#echo "Version: $1" >> ${charttmpdir}/Chart.yaml

mkdir -p ${chartdir}/alameda-operator/ ${chartdir}/admission-controller/ ${chartdir}/alameda-ai/ ${chartdir}/alameda-datahub/ ${chartdir}/alameda-evictioner/

# generate Alameda manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda --namespace $NAMESPACE -f ${values} helm/alameda --output-dir ${tmpoutputdir}

mv ${tmpoutputdir}/alameda/templates/* ${chartdir}/alameda-operator/
echo "$namespaceResource" > ${chartdir}/alameda-operator/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/admission-controller/templates/* ${chartdir}/admission-controller/
echo "$namespaceResource" > ${chartdir}/admission-controller/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/alameda-ai/templates/* ${chartdir}/alameda-ai/
echo "$namespaceResource" > ${chartdir}/alameda-ai/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/datahub/templates/* ${chartdir}/alameda-datahub/
echo "$namespaceResource" > ${chartdir}/alameda-datahub/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/evictioner/templates/* ${chartdir}/alameda-evictioner/
echo "$namespaceResource" > ${chartdir}/alameda-evictioner/namespace.yaml

# generate InfluxDB manifests
mkdir -p ${tmpoutputdir}
helm fetch stable/influxdb --version 1.1.3 --untar --untardir ${tmpchartdir}
helm template --name alameda-influxdb --namespace $NAMESPACE ${tmpchartdir}/influxdb --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-influxdb/
mv ${tmpoutputdir}/influxdb/templates/* ${chartdir}/alameda-influxdb/
echo "$namespaceResource" > ${chartdir}/alameda-influxdb/namespace.yaml

# generate grafana manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda-grafana --namespace $NAMESPACE ./helm/grafana/ --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-grafana/
mv ${tmpoutputdir}/grafana/templates/* ${chartdir}/alameda-grafana/
echo "$namespaceResource" > ${chartdir}/alameda-grafana/namespace.yaml

rm -rf ${tmpdir}

