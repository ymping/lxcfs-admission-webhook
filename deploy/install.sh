#!/usr/bin/env bash

set -e

usage() {
  cat <<EOF
By create lxcfs daemonset and k8s dynamic admission webhook
to help see right container's limitations in the container.

This script will:
1. create deployment of lxcfs daemonset
2. create admission webhook cert stored in a k8s secret
3. create deployment of dynamic admission webhook for patch lxcfs volume for container
4. create k8s MutatingWebhookConfiguration

about lxcfs:
https://linuxcontainers.org/lxcfs/introduction/

about k8s dynamic admission webhook:
https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/

how to generate certificate:
https://kubernetes.io/docs/tasks/administer-cluster/certificates/#openssl for

usage: ${0} [OPTIONS]

The following flags are required.

  --namespace     Namespace where webhook service, lxcfs daemonset and secret reside, default: lxcfs

EOF
  exit 1
}

pre_check() {
  if ! command -v openssl >/dev/null; then
    echo "openssl not found"
    exit 1
  fi
  if ! command -v kubectl >/dev/null; then
    echo "kubectl not found"
    exit 1
  fi
  if ! kubectl cluster-info; then
    echo "Can't connect to kubernetes control plane"
    exit 1
  fi
}

args_parse() {
  while [[ $# -gt 0 ]]; do
    case ${1} in
    --namespace)
      NAMESPACE="$2"
      shift
      ;;
    *)
      usage
      ;;
    esac
    shift
  done

  NAMESPACE=${NAMESPACE:-"lxcfs"}

  local INSTALL_NAME=${INSTALL_NAME:-"lxcfs-admission-webhook"}
  SERVICE=${INSTALL_NAME}
  SECRET_NAME=${INSTALL_NAME}

  cat <<EOF
Create following k8s object in namespace: ${NAMESPACE}:
  webhook service: ${SERVICE}
  webhook secret: ${SECRET_NAME}
  webhook deployment: ${INSTALL_NAME}
  lxcfs daemonset: lxcfs-ds
  mutating webhook configuration: ${INSTALL_NAME}
EOF

}

create_self_signed_cert() {
  # gen certs doc: https://kubernetes.io/docs/tasks/administer-cluster/certificates/#openssl
  echo "creating certs in temp directory: ${TEMP_DIR} "

  local BITS=${BITS:-"2048"}
  local DAYS=${DAYS:-"10950"} # 30 years
  cat <<EOF >"${TEMP_DIR}"/csr.conf
[req]
default_bits = ${BITS}
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
C = CN
ST = Sichuan
L = Chengdu
O = Kubernetes
OU = Dynamic Admission Control
CN=${SERVICE}.${NAMESPACE}.svc

[ req_ext ]
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SERVICE}
DNS.2 = ${SERVICE}.${NAMESPACE}
DNS.3 = ${SERVICE}.${NAMESPACE}.svc
DNS.4 = ${SERVICE}.${NAMESPACE}.svc.cluster
DNS.5 = ${SERVICE}.${NAMESPACE}.svc.cluster.local

[ v3_ext ]
authorityKeyIdentifier = keyid,issuer:always
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

EOF

  # gen ca cert
  openssl genrsa -out "${TEMP_DIR}"/ca-key.pem "${BITS}"
  openssl req -x509 -new -nodes -days "${DAYS}" -key "${TEMP_DIR}"/ca-key.pem -subj "/CN=Kubernetes Admin" -out "${TEMP_DIR}"/ca-cert.pem
  # gen server cert
  openssl genrsa -out "${TEMP_DIR}"/server-key.pem "${BITS}"
  openssl req -new -key "${TEMP_DIR}"/server-key.pem -config "${TEMP_DIR}"/csr.conf -out "${TEMP_DIR}"/server.csr
  openssl x509 -req -in "${TEMP_DIR}"/server.csr -CA "${TEMP_DIR}"/ca-cert.pem -CAkey "${TEMP_DIR}"/ca-key.pem \
    -CAcreateserial -days "${DAYS}" -extensions v3_ext -extfile "${TEMP_DIR}"/csr.conf \
    -out "${TEMP_DIR}"/server-cert.pem

  # create the secret with CA cert and server cert/key
  kubectl create secret generic "${SECRET_NAME}" -n "${NAMESPACE}" \
    --from-file=tls.key="${TEMP_DIR}"/server-key.pem \
    --from-file=tls.crt="${TEMP_DIR}"/server-cert.pem \
    --dry-run=client -o yaml |
    kubectl -n "${NAMESPACE}" apply -f -
}

main() {
  TEMP_DIR=$(mktemp -d -p "$PWD")

  args_parse "$@"
  pre_check

  # 1 Deploy lxcfs daemonset
  kubectl create -n "${NAMESPACE}" -f lxcfs-daemonset.yaml -o yaml --dry-run=client | kubectl -n "${NAMESPACE}" apply -f -

  # 2 Create admission webhook cert
  create_self_signed_cert

  # 3 Deploy admission webhook
  kubectl create -n "${NAMESPACE}" -f deployment.yaml -o yaml --dry-run=client | kubectl -n "${NAMESPACE}" apply -f -
  kubectl create -n "${NAMESPACE}" -f service.yaml -o yaml --dry-run=client | kubectl -n "${NAMESPACE}" apply -f -

  # 4 create k8s MutatingWebhookConfiguration
  CA_BUNDLE=$(base64 <"${TEMP_DIR}"/ca-cert.pem | tr -d '\n')
  export CA_BUNDLE
  export NAMESPACE
  envsubst <"$PWD"/mutatingwebhook.tpl.yaml | kubectl create -o yaml --dry-run=client -f - | kubectl apply -f -
}

main "$@"
