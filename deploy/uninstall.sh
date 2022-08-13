#!/usr/bin/env bash

set -eo pipefail

usage() {
  cat <<EOF
Uninstall lxcfs daemonset and k8s dynamic admission webhook..

This script will:
1. delete k8s MutatingWebhookConfiguration
2. delete deployment of dynamic admission webhook for patch lxcfs volume for container
3. delete admission webhook cert stored in a k8s secret
4. delete deployment of lxcfs daemonset

usage: ${0} [OPTIONS]

The following flags are required.

  --namespace     Namespace where webhook service, lxcfs daemonset and secret reside, default: lxcfs

EOF
  exit 1
}

pre_check() {
  if ! command -v kubectl >/dev/null; then
    echo "kubectl not found"
    exit 1
  fi
  if ! kubectl cluster-info; then
    echo "Can't connect to kubernetes control plane"
    exit 1
  fi
  if ! kubectl get namespaces "${NAMESPACE}"; then
    echo "namespaces: ${NAMESPACE} not exist"
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
      exit 0
      ;;
    esac
    shift
  done

  NAMESPACE=${NAMESPACE:-"lxcfs"}

  INSTALL_NAME=${INSTALL_NAME:-"lxcfs-admission-webhook"}
}

uninstall() {
  kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io "${INSTALL_NAME}"
  kubectl delete -n "${NAMESPACE}" services "${INSTALL_NAME}"
  kubectl delete -n "${NAMESPACE}" deployments.apps "${INSTALL_NAME}"
  kubectl delete -n "${NAMESPACE}" secrets "${INSTALL_NAME}"
  kubectl delete -n "${NAMESPACE}" daemonsets.apps lxcfs-ds
}

main() {
  args_parse "$@"
  pre_check

  cat <<EOF
Delete following k8s object in namespace: ${NAMESPACE}:
  webhook service: ${INSTALL_NAME}
  webhook secret: ${INSTALL_NAME}
  webhook deployment: ${INSTALL_NAME}
  lxcfs daemonset: lxcfs-ds
  mutating webhook configuration: ${INSTALL_NAME}
EOF

  uninstall
}

main "$@"
