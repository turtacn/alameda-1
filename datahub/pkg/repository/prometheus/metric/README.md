# 调试
- 注意node 指标 和 pod 指标的标签在prometheus是否有效
-
- prometheus 服务暴露,便于调试

```yaml
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2021-03-26T10:19:44Z"
  labels:
    operated-prometheus: "true"
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:ownerReferences:
          .: {}
          k:{"uid":"baa1aa4d-0b44-4698-afc4-155ec4e43233"}:
            .: {}
            f:apiVersion: {}
            f:kind: {}
            f:name: {}
            f:uid: {}
    manager: operator
    operation: Update
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:operated-prometheus: {}
      f:spec:
        f:clusterIP: {}
        f:ports:
          .: {}
          k:{"port":9090,"protocol":"TCP"}:
            .: {}
            f:name: {}
            f:port: {}
            f:protocol: {}
            f:targetPort: {}
        f:selector:
          .: {}
          f:app: {}
        f:sessionAffinity: {}
        f:type: {}
    manager: kubectl
    operation: Update
    time: "2021-03-26T10:19:44Z"
  name: prometheus-prometheus-oper-prometheus-nodeport
  namespace: monitoring
spec:
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: web
    nodePort: 32090
  selector:
    app: prometheus
  sessionAffinity: None
  type: NodePort
status:
  loadBalancer: {}
```

