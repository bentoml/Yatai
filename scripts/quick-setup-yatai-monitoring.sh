#!/usr/bin/env bash

set -e

function randstr() {
  LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20
}

# check if jq command exists
if ! command -v jq &> /dev/null; then
  arch=$(uname -m)
  # download jq from github by different arch
  if [[ $arch == "x86_64" && $OSTYPE == 'darwin'* ]]; then
    jq_archived_name="gojq_v0.12.9_darwin_amd64"
  elif [[ $arch == "arm64" && $OSTYPE == 'darwin'* ]]; then
    jq_archived_name="gojq_v0.12.9_darwin_arm64"
  elif [[ $arch == "x86_64" && $OSTYPE == 'linux'* ]]; then
    jq_archived_name="gojq_v0.12.9_linux_amd64"
  elif [[ $arch == "aarch64" && $OSTYPE == 'linux'* ]]; then
    jq_archived_name="gojq_v0.12.9_linux_arm64"
  else
    echo "jq command not found, please install it first"
    exit 1
  fi
  echo "📥 downloading jq from github"
  if [[ $OSTYPE == 'darwin'* ]]; then
    curl -sL -o /tmp/yatai-jq.zip "https://github.com/itchyny/gojq/releases/download/v0.12.9/${jq_archived_name}.zip"
    echo "✅ downloaded jq to /tmp/yatai-jq.zip"
    echo "📦 extracting yatai-jq.zip"
    unzip -q /tmp/yatai-jq.zip -d /tmp
  else
    curl -sL -o /tmp/yatai-jq.tar.gz "https://github.com/itchyny/gojq/releases/download/v0.12.9/${jq_archived_name}.tar.gz"
    echo "✅ downloaded jq to /tmp/yatai-jq.tar.gz"
    echo "📦 extracting yatai-jq.tar.gz"
    tar zxf /tmp/yatai-jq.tar.gz -C /tmp
  fi
  echo "✅ extracted jq to /tmp/${jq_archived_name}"
  jq="/tmp/${jq_archived_name}/gojq"
else
  jq=$(which jq)
fi

# check if kubectl command exists
if ! command -v kubectl >/dev/null 2>&1; then
  echo "😱 kubectl command is not found, please install it first!" >&2
  exit 1
fi

KUBE_VERSION=$(kubectl version --output=json | $jq '.serverVersion.minor')
if [ ${KUBE_VERSION:1:2} -lt 20 ]; then
  echo "😱 install requires at least Kubernetes 1.20" >&2
  exit 1
fi

# check if helm command exists
if ! command -v helm >/dev/null 2>&1; then
  echo "😱 helm command is not found, please install it first!" >&2
  exit 1
fi

namespace=yatai-monitoring

# check if namespace exists
if ! kubectl get namespace ${namespace} >/dev/null 2>&1; then
  echo "🤖 creating namespace ${namespace}"
  kubectl create namespace ${namespace}
  echo "✅ created namespace ${namespace}"
fi

echo "🤖 installing prometheus operator CRDs"
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
kubectl apply --server-side -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/v0.60.1/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml

echo "⏳ waiting for prometheus-operator CRDs to be established..."
kubectl wait --for condition=established --timeout=120s crd/prometheuses.monitoring.coreos.com
kubectl wait --for condition=established --timeout=120s crd/servicemonitors.monitoring.coreos.com
echo "✅ prometheus-operator CRDs are established"

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
echo "🤖 installing prometheus-operator..."
cat <<EOF | helm upgrade --install prometheus prometheus-community/kube-prometheus-stack -n ${namespace} -f -
grafana:
  enabled: false
  forceDeployDatasources: true
  forceDeployDashboards: true
EOF

echo "⏳ waiting for prometheus-operator to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l release=prometheus
echo "✅ prometheus-operator is ready"

echo "🧪 verify that the Prometheus service is running..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/instance=prometheus-kube-prometheus-prometheus
echo "✅ Prometheus service is running"

echo "🧪 verify that the Alertmanager service is running..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/instance=prometheus-kube-prometheus-alertmanager
echo "✅ Alertmanager service is running"

grafana_namespace=yatai-logging

if [ -z "$(kubectl -n ${grafana_namespace} get deploy -l app.kubernetes.io/name=grafana 2>/dev/null)" ]; then
  grafana_namespace=${namespace}
fi

# if grafana namespace is ${namespace} then install grafana
if [ "${grafana_namespace}" = "${namespace}" ]; then
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo update grafana
  echo "🤖 installing Grafana..."
  if ! kubectl -n ${grafana_namespace} get secret grafana > /dev/null 2>&1; then
    grafana_admin_password=$(randstr)
  else
    grafana_admin_password=$(kubectl -n ${grafana_namespace} get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d)
  fi
  cat <<EOF | helm upgrade --install grafana grafana/grafana -n ${grafana_namespace} -f -
adminUser: admin
adminPassword: ${grafana_admin_password}
persistence:
  enabled: true
sidecar:
  dashboards:
    enabled: true
    searchNamespace: ALL
  datasources:
    enabled: true
    searchNamespace: ALL
  notifiers:
    enabled: true
    searchNamespace: ALL
EOF
fi

echo "🧪 verify that the Grafana service is running..."
kubectl -n ${grafana_namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=grafana
echo "✅ Grafana service is running"

echo "🤖 creating PodMonitor for BentoDeployments..."
kubectl apply -f https://raw.githubusercontent.com/bentoml/yatai/main/scripts/monitoring/bentodeployment-podmonitor.yaml
echo "✅ PodMonitor for BentoDeployments is created"

echo "🤖 downloading the BentoDeployment Grafana dashboard json file..."
curl -L https://raw.githubusercontent.com/bentoml/yatai/main/scripts/monitoring/bentodeployment-dashboard.json -o /tmp/bentodeployment-dashboard.json
echo "✅ BentoDeployment Grafana dashboard is downloaded"

echo "🤖 importing the BentoDeployment Grafana dashboard..."
kubectl -n ${grafana_namespace} create configmap bentodeployment-dashboard --from-file=/tmp/bentodeployment-dashboard.json -o yaml --dry-run=client | kubectl apply -f -
kubectl -n ${grafana_namespace} label configmap bentodeployment-dashboard grafana_dashboard=1 --overwrite
echo "✅ BentoDeployment Grafana dashboard is imported"

echo "🤖 creating PodMonitor for BentoFunctions..."
kubectl apply -f https://raw.githubusercontent.com/bentoml/yatai/main/scripts/monitoring/bentofunction-podmonitor.yaml
echo "✅ PodMonitor for BentoFunctions is created"

echo "🤖 downloading the BentoFunction Grafana dashboard json file..."
curl -L https://raw.githubusercontent.com/bentoml/yatai/main/scripts/monitoring/bentofunction-dashboard.json -o /tmp/bentofunction-dashboard.json
echo "✅ BentoFunction Grafana dashboard is downloaded"

echo "🤖 importing the BentoFunction Grafana dashboard..."
kubectl -n ${grafana_namespace} create configmap bentofunction-dashboard --from-file=/tmp/bentofunction-dashboard.json -o yaml --dry-run=client | kubectl apply -f -
kubectl -n ${grafana_namespace} label configmap bentofunction-dashboard grafana_dashboard=1 --overwrite
echo "✅ BentoFunction Grafana dashboard is imported"

echo "🌐 port-forwarding Grafana..."
lsof -i :8888 | tail -n +2 | awk '{print $2}' | xargs -I{} kill {} 2> /dev/null || true
kubectl -n ${grafana_namespace} port-forward svc/grafana 8888:80 --address 0.0.0.0 &
echo "✅ Grafana dashboard is available at: http://localhost:8888/d/TJ3FhiG4z/bentodeployment?orgId=1"

echo "Grafana username: "$(kubectl -n ${grafana_namespace} get secret grafana -o jsonpath='{.data.admin-user}' | base64 -d)
echo "Grafana password: "$(kubectl -n ${grafana_namespace} get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d)
