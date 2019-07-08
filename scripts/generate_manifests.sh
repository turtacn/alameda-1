#!/usr/bin/env bash

# requires helm to be installed
cd `pwd`/$(dirname $BASH_SOURCE)/..
REPO_ROOT_DIR=`pwd`
chartdir=$REPO_ROOT_DIR/example/deployment/kubernetes/
values=$REPO_ROOT_DIR/helm/alameda/values.yaml

while test $# -gt 0;do
    case "$1" in
        -h|--help)
            echo "-f, --values=VALUES     specify values file to use"
            echo "-o, --output-dir=DIR    specify a folder to output result"
            exit 0
            ;;
	-f|--values)
            shift
            if test $# -gt 0;then
                values=$1
            else
                echo "please specify the correct values file"
                exit 1
            fi
            shift
            ;;
        -o|--output-dir)
            shift
            if test $# -gt 0;then
                chartdir=$1
            else
                echo "please specify the correct output folder"
                exit 1
            fi
            shift
            ;;
	*)
            break
            ;;
    esac
done

INFLUXDB_NS_VALUE='value: "{{ .Values.global.component.datahub.influxdbConfig.scheme }}:\/\/{{ .Values.global.component.datahub.influxdbConfig.svcName }}.{{ .Release.Namespace }}.svc:{{ .Values.global.component.datahub.influxdbConfig.port }}"'
INFLUXDB_DEFAULT_NS_VALUE='value: "{{ .Values.global.component.datahub.influxdbConfig.scheme }}:\/\/{{ .Values.global.component.datahub.influxdbConfig.url }}:{{ .Values.global.component.datahub.influxdbConfig.port }}"'
DATAHUB_DEPLOYMENT_FILE=$REPO_ROOT_DIR/helm/alameda/charts/datahub/templates/deployment.yaml
INFLUXDB_NS_URL='url: http:\/\/alameda-influxdb.alameda.svc:8086'
INFLUXDB_DEFAULT_NS_URL='url: http:\/\/alameda-influxdb.default.svc:8086'
GRAFANA_VALUES_FILE=$REPO_ROOT_DIR/helm/grafana/values.yaml

replace::influxdb::ns::to::default() {
    sed -i -e 's/'"$INFLUXDB_NS_VALUE"'/'"$INFLUXDB_DEFAULT_NS_VALUE"'/g' $DATAHUB_DEPLOYMENT_FILE
    sed -i -e 's/'"$INFLUXDB_NS_URL"'/'"$INFLUXDB_DEFAULT_NS_URL"'/g' $GRAFANA_VALUES_FILE
}

restore::influxdb::ns() {
    sed -i -e 's/'"$INFLUXDB_DEFAULT_NS_VALUE"'/'"$INFLUXDB_NS_VALUE"'/g' $DATAHUB_DEPLOYMENT_FILE
    sed -i -e 's/'"$INFLUXDB_DEFAULT_NS_URL"'/'"$INFLUXDB_NS_URL"'/g' $GRAFANA_VALUES_FILE
}

# influxdb helm template does not use release name as namespace,
# the namespace of output manifest is default so we substitute
# influxdb namespace used by other components to default first.
replace::influxdb::ns::to::default

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

mkdir -p ${chartdir}/alameda-operator/ ${chartdir}/admission-controller/ \
    ${chartdir}/alameda-ai/ ${chartdir}/alameda-datahub/ ${chartdir}/alameda-evictioner/ \
    ${chartdir}/alameda-recommender/ ${chartdir}/alameda-executor/ \
    ${chartdir}/alameda-ai-dispatcher/
# isolate rabbitmq component due to it is controlled by predictQueueEnable flag
mkdir -p ${chartdir}/alameda-rabbitmq/

# generate Alameda manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda --namespace $NAMESPACE -f ${values} $REPO_ROOT_DIR/helm/alameda --output-dir ${tmpoutputdir}

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
mv ${tmpoutputdir}/alameda/charts/recommender/templates/* ${chartdir}/alameda-recommender/
echo "$namespaceResource" > ${chartdir}/alameda-recommender/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/executor/templates/* ${chartdir}/alameda-executor/
echo "$namespaceResource" > ${chartdir}/alameda-executor/namespace.yaml

# predictQueueEnable is false, no files can be moved in ai-dispatcher templates
mv ${tmpoutputdir}/alameda/charts/ai-dispatcher/templates/* ${chartdir}/alameda-ai-dispatcher/
echo "$namespaceResource" > ${chartdir}/alameda-ai-dispatcher/namespace.yaml
mv ${tmpoutputdir}/alameda/charts/rabbitmq/templates/* ${chartdir}/alameda-rabbitmq/
echo "$namespaceResource" > ${chartdir}/alameda-rabbitmq/namespace.yaml

# generate InfluxDB manifests
mkdir -p ${tmpoutputdir}
helm fetch stable/influxdb --version 1.1.9 --untar --untardir ${tmpchartdir}
helm template --name alameda-influxdb --namespace $NAMESPACE ${tmpchartdir}/influxdb --set persistence.enabled=false --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-influxdb/
mv ${tmpoutputdir}/influxdb/templates/* ${chartdir}/alameda-influxdb/
echo "$namespaceResource" > ${chartdir}/alameda-influxdb/namespace.yaml

# generate grafana manifests
mkdir -p ${tmpoutputdir}
helm template --name alameda-grafana --namespace $NAMESPACE $REPO_ROOT_DIR/helm/grafana/ --output-dir ${tmpoutputdir}

mkdir -p ${chartdir}/alameda-grafana/
mv ${tmpoutputdir}/grafana/templates/* ${chartdir}/alameda-grafana/
echo "$namespaceResource" > ${chartdir}/alameda-grafana/namespace.yaml

# generate rabbitmq manifests
#mkdir -p ${tmpoutputdir}
#helm template --name alameda-rabbitmq --namespace $NAMESPACE $REPO_ROOT_DIR/helm/alameda/charts/rabbitmq/ --output-dir ${tmpoutputdir}

#mkdir -p ${chartdir}/alameda-rabbitmq/
#mv ${tmpoutputdir}/rabbitmq/templates/* ${chartdir}/alameda-rabbitmq/
#echo "$namespaceResource" > ${chartdir}/alameda-rabbitmq/namespace.yaml

rm -rf ${tmpdir}
restore::influxdb::ns
