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


echo "⏳ waiting for minio-operator to be ready..."
if ! kubectl wait --for=condition=ready --timeout=60s pod -l app.kubernetes.io/instance=minio-operator -A; then
  echo "😱 minio-operator is not ready"

  helm repo add minio https://operator.min.io/ || true
  helm repo update minio

  echo "🤖 installing minio-operator..."
  helm upgrade --install minio-operator minio/operator -n ${namespace}

  echo "⏳ waiting for minio-operator to be ready..."
  kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/instance=minio-operator
fi
echo "✅ minio-operator is ready"

minio_secret_name=yatai-logging-minio

# check if logging minio secret not exists
echo "🧐 checking if secret ${minio_secret_name} exists..."
if ! kubectl get secret ${minio_secret_name} -n ${namespace} >/dev/null 2>&1; then
  echo "🥹 secret ${minio_secret_name} not found"

  echo "🤖 creating secret ${minio_secret_name}"
  kubectl create secret generic ${minio_secret_name} \
    --from-literal=accesskey=$(randstr) \
    --from-literal=secretkey=$(randstr) \
    -n ${namespace}
  echo "✅ created secret ${minio_secret_name}"
else
  echo "🤩 secret ${minio_secret_name} already exists"
fi

S3_ENDPOINT=minio.${namespace}.svc.cluster.local
S3_REGION=foo
S3_BUCKET_NAME=loki-data
S3_SECURE=true
S3_ACCESS_KEY=$(kubectl -n ${namespace} get secret ${minio_secret_name} -o jsonpath='{.data.accesskey}' | base64 -d)
S3_SECRET_KEY=$(kubectl -n ${namespace} get secret ${minio_secret_name} -o jsonpath='{.data.secretkey}' | base64 -d)

# check if S3_ACCESS_KEY is empty
if [ -z "$S3_ACCESS_KEY" ]; then
  echo "🥹 S3_ACCESS_KEY is empty" >&2
  exit 1
fi

echo "🤖 make sure has standard storageclass..."
if ! kubectl get storageclass standard >/dev/null 2>&1; then
  echo "😱 standard storageclass not found"
  echo "🤖 creating standard storageclass..."
  # get the default storageclass
  default_storageclass=$(kubectl get storageclass -o json | $jq -r '.items[] | select(.metadata.annotations."storageclass.kubernetes.io/is-default-class" == "true") | .metadata.name')
  if [ -z "$default_storageclass" ]; then
    echo "😱 default storageclass not found"
    exit 1
  fi
  # copy the default storageclass to standard
  echo "🤖 copying default storageclass to standard..."
  kubectl get storageclass ${default_storageclass} -o yaml | sed 's/  name: '"${default_storageclass}"'/  name: standard/' | kubectl apply -f -
  # remove the default annotation for standard storageclass
  kubectl patch storageclass standard -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
  echo "✅ created standard storageclass"
else
  echo "🤩 standard storageclass already exists"
fi

helm repo add minio https://operator.min.io/ || true
helm repo update minio

echo "🤖 creating MinIO Tenant..."
helm upgrade --install yatai-logging-minio-tenant minio/tenant \
  -n ${namespace} \
  --set secrets.accessKey=${S3_ACCESS_KEY} \
  --set secrets.secretKey=${S3_SECRET_KEY} \
  --set tenant.name=yatai-logging-minio

echo "⏳ waiting for minio tenant to be ready..."
# this retry logic is to avoid kubectl wait errors due to minio tenant resources not being created
for i in $(seq 1 10); do
  if kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l v1.min.io/tenant=yatai-logging-minio; then
    echo "✅ minio tenant is ready"
    break
  else
    if [ $i -eq 10 ]; then
      echo "😱 minio tenant is not ready"
      exit 1
    fi
    echo "😱 minio tenant is not ready, retrying..."
    sleep 5
    continue
  fi
done

echo "🧪 testing MinIO connection..."
for i in $(seq 1 10); do
  kubectl -n ${namespace} delete pod s3-client 2> /dev/null || true

  if kubectl run s3-client --rm --tty -i --restart='Never' \
      --namespace ${namespace} \
      --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
      --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
      --image quay.io/bentoml/s3-client:0.0.1 \
      --command -- sh -c "s3-client -e https://$S3_ENDPOINT listbuckets 2>/dev/null"; then
        echo "✅ MinIO connection is successful"
        break
  else
    if [ $i -eq 10 ]; then
      echo "😱 MinIO connection failed"
      exit 1
    fi
    echo "😱 MinIO connection failed, retrying..."
    sleep 5
    continue
  fi
done

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update grafana
echo "🤖 installing Loki..."
cat <<EOF | helm upgrade --install loki grafana/loki-distributed -n ${namespace} --version 0.65.0 -f -
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
        s3: https://$S3_ACCESS_KEY:$S3_SECRET_KEY@$S3_ENDPOINT/$S3_BUCKET_NAME
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
cat <<EOF | helm upgrade --install promtail grafana/promtail --version 6.6.1 -n ${namespace} -f -
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
