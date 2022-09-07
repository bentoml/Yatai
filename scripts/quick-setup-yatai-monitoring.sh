#!/bin/bash

set -e

# check if jq command exists
if ! command -v jq &> /dev/null; then
  # download jq from github by different arch
  case "$(uname -m)" in
    x86_64)
      JQ_ARCH=jq-linux64
      ;;
    aarch64)
      JQ_ARCH=jq-linux64
      ;;
    armv7l)
      JQ_ARCH=jq-linux32
      ;;
    Darwin)
      JQ_ARCH=jq-osx-amd64
      ;;
    *)
      echo "Unsupported architecture $(uname -m)"
      exit 1
      ;;
  esac
  echo "📥 downloading jq from github"
  curl -sL -o /tmp/yatai-jq "https://github.com/stedolan/jq/releases/download/jq-1.6/${JQ_ARCH}"
  echo "✅ downloaded jq to /tmp/yatai-jq"
  chmod +x /tmp/yatai-jq
  jq=/tmp/yatai-jq
else
  jq=$(which jq)
fi

# check if kubectl command exists
if ! command -v kubectl >/dev/null 2>&1; then
  echo "😱 kubectl command is not found, please install it first!" >&2
  exit 1
fi

KUBE_VERSION=$(kubectl version --output=json | jq '.serverVersion.minor')
if [ ${KUBE_VERSION:1:2} -lt 20 ]; then
  echo "😱 install requires at least Kubernetes 1.20" >&2
  exit 1
fi

# check if helm command exists
if ! command -v helm >/dev/null 2>&1; then
  echo "😱 helm command is not found, please install it first!" >&2
  exit 1
fi

kubectl create ns yatai-monitoring

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
echo "🤖 installing prometheus-operator..."
cat <<EOF | helm install prometheus prometheus-community/kube-prometheus-stack -n yatai-monitoring -f -
grafana:
  enabled: false
  forceDeployDatasources: true
  forceDeployDashboards: true
EOF

echo "⏳ waiting for prometheus-operator to be ready..."
kubectl -n yatai-monitoring wait --for=condition=ready --timeout=600s pod -l release=prometheus
echo "✅ prometheus-operator is ready"

echo "⏳ waiting for prometheus-operator CRDs to be established..."
kubectl wait --for condition=established --timeout=120s crd/prometheuses.monitoring.coreos.com
kubectl wait --for condition=established --timeout=120s crd/servicemonitors.monitoring.coreos.com
echo "✅ prometheus-operator CRDs are established"

echo "🧪 verify that the Prometheus service is running..."
kubectl -n yatai-monitoring wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/instance=prometheus-kube-prometheus-prometheus
echo "✅ Prometheus service is running"

echo "🧪 verify that the Alertmanager service is running..."
kubectl -n yatai-monitoring wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/instance=prometheus-kube-prometheus-alertmanager
echo "✅ Alertmanager service is running"

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update grafana
echo "🤖 installing Grafana..."
cat <<EOF | helm install grafana grafana/grafana -n yatai-monitoring -f -
adminUser: admin
adminPassword: $(openssl rand -base64 16)
persistence:
  enabled: true
sidecar:
  dashboards:
    enabled: true
  datasources:
    enabled: true
  notifiers:
    enabled: true
EOF

echo "🧪 verify that the Grafana service is running..."
kubectl -n yatai-monitoring wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=grafana
echo "✅ Grafana service is running"

echo "🤖 creating PodMonitor for BentoDeployments..."
kubectl apply -f https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/monitoring/bentodeployment-podmonitor.yaml
echo "✅ PodMonitor for BentoDeployments is created"

echo "🤖 downloading the BentoDeployment Grafana dashboard json file..."
curl -L https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/monitoring/bentodeployment-dashboard.json -o /tmp/bentodeployment-dashboard.json
echo "✅ BentoDeployment Grafana dashboard is downloaded"

echo "🤖 importing the BentoDeployment Grafana dashboard..."
kubectl -n yatai-monitoring create configmap bentodeployment-dashboard --from-file=/tmp/bentodeployment-dashboard.json
kubectl -n yatai-monitoring label configmap bentodeployment-dashboard grafana_dashboard=1
echo "✅ BentoDeployment Grafana dashboard is imported"

echo "🌐 port-forwarding Grafana..."
kubectl -n yatai-monitoring port-forward svc/grafana 8888:80 --address 0.0.0.0 &
echo "✅ Grafana dashboard is available at: http://localhost:8888/d/TJ3FhiG4z/bentodeployment?orgId=1"

echo "Grafana username: "$(kubectl -n yatai-monitoring get secret grafana -o jsonpath='{.data.admin-user}' | base64 -d)
echo "Grafana password: "$(kubectl -n yatai-monitoring get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d)
