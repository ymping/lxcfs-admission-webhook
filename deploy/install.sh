#!/usr/bin/env bash

set -eo pipefail

usage() {
  cat <<EOF
usage: ${0} [advanced options]

advanced options:
defined resources name:
  --namespace         kubernetes namespace where webhook service, lxcfs daemonset and secret reside, default: lxcfs
  --deployment        webhook deployment name, default: lxcfs-admission-webhook
  --service           LXCFS admission webhook service name, default: lxcfs-admission-webhook
  --secret            LXCFS admission webhook mutating secret name, default: lxcfs-admission-webhook
  --daemonset         LXCFS daemonset name, default: lxcfs-ds
  --mutating          mutating admission name, default: lxcfs-admission-webhook

  --create-cert-only  generate a self-signed certificate in current directory

EOF
}

help() {
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
https://kubernetes.io/docs/tasks/administer-cluster/certificates/#openssl

EOF

  usage
}

pre_check() {
  if ! kubectl cluster-info; then
    echo "Can't connect to kubernetes control plane"
    exit 101
  fi
}

create_self_signed_cert() {
  # gen certs doc: https://kubernetes.io/docs/tasks/administer-cluster/certificates/#openssl
  echo "Creating certs in directory: ${CERT_DIR} "

  local BITS=${BITS:-"2048"}
  local DAYS=${DAYS:-"10950"} # 30 years
  cat <<EOF >"${CERT_DIR}"/csr.conf
[ req ]
default_bits = ${BITS}
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = CN
ST = Sichuan
L = Chengdu
O = Kubernetes
OU = Dynamic Admission Control
CN=${WH_SVC}.${NAMESPACE}.svc

[ req_ext ]
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${WH_SVC}
DNS.2 = ${WH_SVC}.${NAMESPACE}
DNS.3 = ${WH_SVC}.${NAMESPACE}.svc
DNS.4 = ${WH_SVC}.${NAMESPACE}.svc.cluster
DNS.5 = ${WH_SVC}.${NAMESPACE}.svc.cluster.local

[ v3_ext ]
authorityKeyIdentifier = keyid,issuer:always
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

EOF

  # gen ca cert
  openssl genrsa -out "${CERT_DIR}"/ca-key.pem "${BITS}"
  openssl req -x509 -new -nodes -days "${DAYS}" -key "${CERT_DIR}"/ca-key.pem -subj "/CN=Kubernetes Admin" -out "${CERT_DIR}"/ca-cert.pem
  # gen server cert
  openssl genrsa -out "${CERT_DIR}"/server-key.pem "${BITS}"
  openssl req -new -key "${CERT_DIR}"/server-key.pem -config "${CERT_DIR}"/csr.conf -out "${CERT_DIR}"/server.csr
  openssl x509 -req -in "${CERT_DIR}"/server.csr -CA "${CERT_DIR}"/ca-cert.pem -CAkey "${CERT_DIR}"/ca-key.pem \
    -CAcreateserial -days "${DAYS}" -extensions v3_ext -extfile "${CERT_DIR}"/csr.conf \
    -out "${CERT_DIR}"/server-cert.pem
}

create_k8s_resources() {
  export NAMESPACE
  export WH_DEP
  export WH_SVC
  export WH_SECRET
  export MUTATING_WH_CONFIG
  export LXCFS_DS

  # 1 Deploy lxcfs daemonset
  envsubst <"$PWD"/lxcfs-daemonset.tpl.yaml | kubectl create -n "${NAMESPACE}" -o yaml --dry-run=client -f - | kubectl -n "${NAMESPACE}" apply -f -

  # 2 Create admission webhook cert
  CERT_DIR=$(mktemp -d)
  create_self_signed_cert
  kubectl create secret generic "${WH_SECRET}" -n "${NAMESPACE}" \
    --from-file=tls.key="${CERT_DIR}"/server-key.pem \
    --from-file=tls.crt="${CERT_DIR}"/server-cert.pem \
    --dry-run=client -o yaml |
    kubectl -n "${NAMESPACE}" apply -f -

  # 3 Deploy admission webhook
  envsubst <"$PWD"/deployment.tpl.yaml | kubectl create -n "${NAMESPACE}" -o yaml --dry-run=client -f - | kubectl -n "${NAMESPACE}" apply -f -
  envsubst <"$PWD"/service.tpl.yaml | kubectl create -n "${NAMESPACE}" -o yaml --dry-run=client -f - | kubectl -n "${NAMESPACE}" apply -f -

  # 4 Create k8s MutatingWebhookConfiguration
  CA_BUNDLE=$(base64 <"${CERT_DIR}"/ca-cert.pem | tr -d '\n')
  export CA_BUNDLE
  envsubst <"$PWD"/mutatingwebhook.tpl.yaml | kubectl create -o yaml --dry-run=client -f - | kubectl apply -f -
}

main() {
  # set resources default name
  NAMESPACE=lxcfs
  WH_DEP=lxcfs-admission-webhook
  WH_SVC=lxcfs-admission-webhook
  WH_SECRET=lxcfs-admission-webhook
  MUTATING_WH_CONFIG=lxcfs-admission-webhook
  LXCFS_DS=lxcfs-ds
  CREATE_CERT_ONLY=false

  if [[ $# -ge 1 ]]; then
    case $1 in
    --namespace)
      NAMESPACE=${2:-NAMESPACE}
      shift 2
      ;;
    --deployment)
      WH_DEP=${2:-WH_DEP}
      shift 2
      ;;
    --service)
      WH_SVC=${2:-WH_SVC}
      shift 2
      ;;
    --secret)
      WH_SECRET=${2:-WH_SECRET}
      shift 2
      ;;
    --mutating)
      MUTATING_WH_CONFIG=${2:-MUTATING_WH_CONFIG}
      shift 2
      ;;
    --daemonset)
      LXCFS_DS=${2:-LXCFS_DS}
      shift 2
      ;;
    --create-cert-only)
      CREATE_CERT_ONLY=true
      shift
      ;;
    --help | -h)
      help
      exit 0
      ;;
    *)
      echo -e "unknown parameter：$1\n"
      usage
      exit 22
      ;;
    esac
  fi

  # just create cert and exit if flag CREATE_CERT_ONLY set to true
  if [[ ${CREATE_CERT_ONLY} == true ]]; then
    CERT_DIR="${PWD}/certs"
    mkdir -p "${CERT_DIR}"
    create_self_signed_cert
    exit 0
  fi

  pre_check

  cat <<EOF
Create following k8s resources in namespace: ${NAMESPACE}
  webhook service: ${WH_SVC}
  webhook secret: ${WH_SECRET}
  webhook deployment: ${WH_DEP}
  lxcfs daemonset: ${LXCFS_DS}
  mutating webhook configuration: ${MUTATING_WH_CONFIG}
EOF

  create_k8s_resources
}

main "$@"
