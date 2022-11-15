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

uninstall() {
  cat <<EOF
Delete following k8s object in namespace: ${NAMESPACE}:
  webhook deployment: ${WH_DEP}
  webhook service: ${WH_SVC}
  webhook secret: ${WH_SECRET}
  lxcfs daemonset: ${LXCFS_DS}
  mutating webhook configuration: ${MUTATING_WH_CONFIG}
EOF

  kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io "${MUTATING_WH_CONFIG}"
  kubectl delete -n "${NAMESPACE}" services "${WH_SVC}"
  kubectl delete -n "${NAMESPACE}" deployments.apps "${WH_DEP}"
  kubectl delete -n "${NAMESPACE}" secrets "${WH_SECRET}"
  kubectl delete -n "${NAMESPACE}" daemonsets.apps "${LXCFS_DS}"
}

main() {
  # set resources default name
  NAMESPACE=lxcfs
  WH_DEP=lxcfs-admission-webhook
  WH_SVC=lxcfs-admission-webhook
  WH_SECRET=lxcfs-admission-webhook
  MUTATING_WH_CONFIG=lxcfs-admission-webhook
  LXCFS_DS=lxcfs-ds

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
    *)
      echo -e "unknown parameter：$1\n"
      usage
      exit 22
      ;;
    esac
  fi

  pre_check

  uninstall
}

main "$@"
