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

tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'charttmpdir'`
tmpchartdir=${tmpdir}
tmpoutputdir=${tmpdir}/output

#cp -R helm/alameda ${tmpchartdir}
#echo "Version: $1" >> ${charttmpdir}/Chart.yaml

mkdir -p ${chartdir}/alameda-operator/ ${chartdir}/admission-controller/ ${chartdir}/alameda-ai/ ${chartdir}/alameda-datahub/ ${chartdir}/alameda-evictioner/

# generate Alameda manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda --namespace alameda -f ${values} helm/alameda --output-dir ${tmpoutputdir}

mv ${tmpoutputdir}/alameda/templates/* ${chartdir}/alameda-operator/
mv ${tmpoutputdir}/alameda/charts/admission-controller/templates/* ${chartdir}/admission-controller/
mv ${tmpoutputdir}/alameda/charts/alameda-ai/templates/* ${chartdir}/alameda-ai/
mv ${tmpoutputdir}/alameda/charts/datahub/templates/* ${chartdir}/alameda-datahub/
mv ${tmpoutputdir}/alameda/charts/evictioner/templates/* ${chartdir}/alameda-evictioner/

# generate InfluxDB manifests
mkdir -p ${tmpoutputdir}
helm fetch stable/influxdb --version 1.1.3 --untar --untardir ${tmpchartdir}
helm template --name alameda-influxdb --namespace alameda ${tmpchartdir}/influxdb --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-influxdb/
mv ${tmpoutputdir}/influxdb/templates/* ${chartdir}/alameda-influxdb/

# generate grafana manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda-grafana --namespace alameda ./helm/grafana/ --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-grafana/
mv ${tmpoutputdir}/grafana/templates/* ${chartdir}/alameda-grafana/

rm -rf ${tmpdir}
