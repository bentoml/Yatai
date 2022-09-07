#!/usr/bin/env bash

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

# check if yatai-system namespace exists
if ! kubectl get namespace yatai-system >/dev/null 2>&1; then
  echo "😱 yatai-system namespace is not found, please install Yatai first!" >&2
  exit 1
fi

namespace=yatai-logging

# check if ${namespace} namespace exists
if ! kubectl get namespace ${namespace} >/dev/null 2>&1; then
  echo "🤖 creating namespace ${namespace}"
  kubectl create namespace ${namespace}
  echo "✅ created namespace ${namespace}"
fi

logging_minio_secret_name=yatai-logging-minio

# check if logging minio secret not exists
if ! kubectl get secret ${logging_minio_secret_name} -n ${namespace} >/dev/null 2>&1; then
  echo "🤖 creating secret ${logging_minio_secret_name}"
  kubectl create secret generic ${logging_minio_secret_name} \
    --from-literal=accesskey=$(echo $RANDOM | md5sum | head -c 20; echo -n) \
    --from-literal=secretkey=$(echo $RANDOM | md5sum | head -c 20; echo -n) \
    -n ${namespace}
  echo "✅ created secret ${logging_minio_secret_name}"
fi

echo "🤖 creating MinIO Tenant..."
cat <<EOF | kubectl apply -f -
apiVersion: minio.min.io/v2
kind: Tenant
metadata:
  labels:
    app: minio
  name: logging-minio
  namespace: ${namespace}
spec:
  credsSecret:
    name: ${logging_minio_secret_name}
  image: quay.io/bentoml/minio-minio:RELEASE.2021-10-06T23-36-31Z
  imagePullPolicy: IfNotPresent
  mountPath: /export
  podManagementPolicy: Parallel
  pools:
  - servers: 4
    volumeClaimTemplate:
      metadata:
        name: data
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 20Gi
    volumesPerServer: 4
  requestAutoCert: false
  s3:
    bucketDNS: false
  subPath: /data
EOF

echo "⏳ waiting for minio tenant to be ready..."
for i in $(seq 1 10); do
  kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app=minio && break || sleep 5
done
echo "✅ minio tenant is ready"

S3_ENDPOINT=minio.${namespace}.svc.cluster.local
S3_REGION=foo
S3_BUCKET_NAME=loki-data
S3_SECURE=false
S3_ACCESS_KEY=$(kubectl -n ${namespace} get secret ${logging_minio_secret_name} -o jsonpath='{.data.accesskey}' | base64 -d)
S3_SECRET_KEY=$(kubectl -n ${namespace} get secret ${logging_minio_secret_name} -o jsonpath='{.data.secretkey}' | base64 -d)

echo "🧪 testing MinIO connection..."
for i in $(seq 1 10); do
  kubectl -n ${namespace} delete pod s3-client 2> /dev/null || true

  kubectl run s3-client --rm --tty -i --restart='Never' \
      --namespace ${namespace} \
      --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
      --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
      --image quay.io/bentoml/s3-client:0.0.1 \
      --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets 2>/dev/null && echo successfully || echo failed" && break || sleep 5
done
echo "✅ MinIO connection is successful"

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update grafana
echo "🤖 installing Loki..."
cat <<EOF | helm upgrade --install loki grafana/loki-distributed -n ${namespace} -f -
loki:
  image:
    registry: quay.io/bentoml
    repository: grafana-loki
    tag: 2.6.1
  structuredConfig:
    ingester:
      # Disable chunk transfer which is not possible with statefulsets
      # and unnecessary for boltdb-shipper
      max_transfer_retries: 0
      chunk_idle_period: 1h
      chunk_target_size: 1536000
      max_chunk_age: 1h
    storage_config:
      aws:
        s3: http://$S3_ACCESS_KEY:$S3_SECRET_KEY@$S3_ENDPOINT/$S3_BUCKET_NAME
        s3forcepathstyle: true
      boltdb_shipper:
        shared_store: s3
    schema_config:
      configs:
        - from: 2020-09-07
          store: boltdb-shipper
          object_store: s3
          schema: v11
          index:
            prefix: loki_index_
            period: 24h
gateway:
  image:
    registry: quay.io/bentoml
    repository: nginxinc-nginx-unprivileged
    tag: 1.19-alpine
EOF


echo "⏳ waiting for Loki to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=loki-distributed
echo "✅ Loki is ready"

echo "🤖 installing Promtail..."
cat <<EOF | helm upgrade --install promtail grafana/promtail -n ${namespace} -f -
config:
  clients:
    - url: http://loki-loki-distributed-gateway.${namespace}.svc.cluster.local/loki/api/v1/push
      tenant_id: 1
  snippets:
    pipelineStages:
      - docker: {}
      - cri: {}
      - multiline:
          firstline: '^[^ ]'
          max_wait_time: 500ms
    extraRelabelConfigs:
      - action: replace
        source_labels:
          - __meta_kubernetes_pod_label_yatai_ai_bento_deployment
        target_label: yatai_bento_deployment
      - action: replace
        source_labels:
          - __meta_kubernetes_pod_label_yatai_ai_bento_deployment_component_type
        target_label: yatai_bento_deployment_component_type
      - action: replace
        source_labels:
          - __meta_kubernetes_pod_label_yatai_ai_bento_deployment_component_name
        target_label: yatai_bento_deployment_component_name
EOF

echo "⏳ waiting for Promtail to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=promtail
echo "✅ Promtail is ready"

grafana_namespace=yatai-monitoring

if [ -z "$(kubectl -n ${grafana_namespace} get deploy -l app.kubernetes.io/name=grafana 2>/dev/null)" ]; then
  grafana_namespace=${namespace}
fi

# if grafana namespace is ${namespace} then install grafana
if [ "${grafana_namespace}" = "${namespace}" ]; then
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo update grafana
  echo "🤖 installing Grafana..."
  if ! kubectl -n ${grafana_namespace} get secret grafana > /dev/null 2>&1; then
    grafana_admin_password=$(openssl rand -base64 16)
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

if [ "${grafana_namespace}" = "${namespace}" ]; then
  echo "🤖 importing Grafana datasource..."
  cat <<EOF > /tmp/loki-datasource.yaml
apiVersion: 1
datasources:
- name: Loki
  type: loki
  access: proxy
  url: http://loki-loki-distributed-gateway.${namespace}.svc.cluster.local
  version: 1
  editable: false
EOF

  kubectl -n ${namespace} create configmap loki-datasource --from-file=/tmp/loki-datasource.yaml -o yaml --dry-run=client | kubectl apply -f -
  kubectl -n ${namespace} label configmap loki-datasource grafana_datasource=1 --overwrite
  echo "✅ Grafana datasource is imported"
fi

echo "🤖 restarting Grafana..."
kubectl -n ${grafana_namespace} rollout restart deployment grafana

echo "⏳ waiting for Grafana to be ready..."
kubectl -n ${grafana_namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=grafana
echo "✅ Grafana is ready"

echo "🌐 port-forwarding Grafana..."
lsof -i :8889 | tail -n +2 | awk '{print $2}' | xargs -I{} kill {} 2> /dev/null || true
kubectl -n ${grafana_namespace} port-forward svc/grafana 8889:80 --address 0.0.0.0 &
echo "✅ Grafana dashboard is available at: http://localhost:8889/explore"

echo "Grafana username: "$(kubectl -n ${grafana_namespace} get secret grafana -o jsonpath='{.data.admin-user}' | base64 -d)
echo "Grafana password: "$(kubectl -n ${grafana_namespace} get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d)
