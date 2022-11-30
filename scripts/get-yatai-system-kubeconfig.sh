#!/usr/bin/env bash

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <filename>"
    exit 1
fi

namespace="yatai-system"
service_account="yatai"

echo "🧪 verifying if namespace ${namespace} already exists..."
if ! kubectl get ns "${namespace}" > /dev/null 2>&1; then
  echo "🤖 create namespace ${namespace}..."
  kubectl create ns ${namespace}
  echo "🥂 namespace ${namespace} created!"
else
  echo "🥂 namespace ${namespace} already exists!"
fi

echo "🧪 verifying if serviceaccount ${service_account} already exists..."
if ! kubectl -n "${namespace}" get sa "${service_account}" > /dev/null 2>&1; then
  echo "🤖 create serviceaccount ${service_account}..."
  kubectl -n "${namespace}" create sa "${service_account}"
  echo "🥂 serviceaccount ${service_account} created!"
else
  echo "🥂 serviceaccount ${service_account} already exists!"
fi

echo "🤖 generating kubeconfig..."

TEMPDIR=$(mktemp -d)

trap "{ rm -rf $TEMPDIR ; exit 255; }" EXIT

SA_SECRET=$(kubectl -n ${namespace} get sa ${service_account} -o=jsonpath='{.secrets[0].name}')

BEARER_TOKEN=$(kubectl get secret -n ${namespace} ${SA_SECRET} -o jsonpath='{.data.token}' | base64 -d )

kubectl -n ${namespace} get secret ${SA_SECRET} -o jsonpath='{.data.ca\.crt}' | base64 -d > $TEMPDIR/ca.crt

CLUSTER_URL=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')

KUBECONFIG="${TEMPDIR}/kubeconfig.yaml"

kubectl config --kubeconfig="${KUBECONFIG}" \
  set-cluster \
  "${CLUSTER_URL}" \
  --server="${CLUSTER_URL}" \
  --certificate-authority="${TEMPDIR}/ca.crt" \
  --embed-certs=true

kubectl config --kubeconfig="${KUBECONFIG}" \
  set-credentials "${service_account}" --token="${BEARER_TOKEN}"

kubectl config --kubeconfig="${KUBECONFIG}" \
  set-context "${service_account}" \
  --cluster="${CLUSTER_URL}" \
  --namespace="${namespace}" \
  --user="${service_account}"

kubectl config --kubeconfig="${KUBECONFIG}" \
  use-context "${service_account}"

cp "${KUBECONFIG}" "${1}"

echo "🥂 kubeconfig dumped to ${1}"
