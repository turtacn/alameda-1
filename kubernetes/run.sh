source ~/turta/misc/kind/kind.rc
kubectl create -f ./0alameda-operator/ -n alameda
kubectl create -f ./0alameda-datahub/ -n alameda
kubectl create -f ./0alameda-ai/ -n alameda 
kubectl create -f ./0alameda-evictioner/ -n alameda
#kubectl create -f ./0admission-controller/ -n alameda
kubectl create -f ./0alameda-influxdb/ -n alameda
kubectl create -f ./0alameda-grafana/ -n alameda
kubectl create -f ./0alameda-ai-dispatcher/ -n alameda
kubectl create -f ./0alameda-executor/ -n alameda
kubectl create -f ./0alameda-notifier/ -n alameda
kubectl create -f ./0alameda-rabbitmq/ -n alameda
kubectl create -f ./0alameda-recommender/ -n alameda
