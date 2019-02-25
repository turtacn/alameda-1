#!/bin/sh

TMP_DIR="/tmp/admission-controller-certs"

mkdir -p ${TMP_DIR}

openssl genrsa -out ${TMP_DIR}/caKey.pem 2048
openssl req -x509 -new -nodes -key ${TMP_DIR}/caKey.pem -days 100000 -out ${TMP_DIR}/caCert.pem -subj "/CN=admission_ca"

cat > ${TMP_DIR}/server.conf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
EOF

openssl genrsa -out ${TMP_DIR}/serverKey.pem 2048
openssl req -new -key ${TMP_DIR}/serverKey.pem -out ${TMP_DIR}/server.csr -subj "/CN=admission-controller.alameda.svc" -config ${TMP_DIR}/server.conf
openssl x509 -req -in ${TMP_DIR}/server.csr -CA ${TMP_DIR}/caCert.pem -CAkey ${TMP_DIR}/caKey.pem -CAcreateserial -out ${TMP_DIR}/serverCert.pem -days 100000 -extensions v3_req -extfile ${TMP_DIR}/server.conf

kubectl create namespace alameda
kubectl create secret --namespace=alameda generic admission-controller-tls-certs --from-file=${TMP_DIR}/caKey.pem --from-file=${TMP_DIR}/caCert.pem --from-file=${TMP_DIR}/serverKey.pem --from-file=${TMP_DIR}/serverCert.pem

cat > ${TMP_DIR}/webhook-configuration.yaml <<EOF
kind: MutatingWebhookConfiguration
apiVersion: admissionregistration.k8s.io/v1beta1
metadata:
  name: mutating-webhook.admission-controller.alameda.svc
webhooks:
  - name: mutating-webhook.admission-controller.alameda.svc
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"] 
    clientConfig:
      caBundle: $(cat ${TMP_DIR}/caCert.pem | base64 | tr -d '\n')
      service:
        namespace: alameda
        name: admission-controller
        path: "/pods"
EOF

kubectl create -f ${TMP_DIR}/webhook-configuration.yaml
rm -rf ${TMP_DIR}
