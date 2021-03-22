kubectl delete -f ./0alameda-operator/ -n alameda
kubectl delete -f ./0alameda-datahub/ -n alameda
kubectl delete -f ./0alameda-ai/ -n alameda 
kubectl delete -f ./0alameda-evictioner/ -n alameda
#kubectl delete -f ./0admission-controller/ -n alameda
kubectl delete -f ./0alameda-influxdb/ -n alameda
kubectl delete -f ./0alameda-grafana/ -n alameda
