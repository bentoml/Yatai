#!/usr/bin/env bash

set -e

DEVEL=${DEVEL:-false}
DEVEL_HELM_REPO=${DEVEL_HELM_REPO:-false}

function randstr() {
  LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20
}

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

namespace=yatai-system

# check if yatai-system namespace exists
if ! kubectl get namespace ${namespace} >/dev/null 2>&1; then
  echo "🤖 creating namespace ${namespace}"
  kubectl create namespace ${namespace}
  echo "✅ created namespace ${namespace}"
fi

if ! kubectl -n ${namespace} get secret postgresql-ha-postgresql >/dev/null 2>&1; then
  postgresql_password=$(randstr)
  repmgr_password=$(randstr)
else
  postgresql_password=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.postgresql-password}" | base64 -d)
  repmgr_password=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.repmgr-password}" | base64 -d)
fi

if ! kubectl -n ${namespace} get secret postgresql-ha-pgpool >/dev/null 2>&1; then
  pgpool_admin_password=$(randstr)
else
  pgpool_admin_password=$(kubectl -n ${namespace} get secret postgresql-ha-pgpool -o jsonpath="{.data.admin-password}" | base64 -d)
fi

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update bitnami
echo "🤖 installing PostgreSQL..."
helm upgrade --install postgresql-ha bitnami/postgresql-ha -n ${namespace} \
  --set postgresql.password=${postgresql_password} \
  --set postgresql.repmgrPassword=${repmgr_password} \
  --set pgpool.adminPassword=${pgpool_admin_password}

echo "⏳ waiting for PostgreSQL to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=postgresql-ha
echo "✅ PostgreSQL is ready"

PG_PASSWORD=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.postgresql-password}" | base64 -d)
PG_HOST=postgresql-ha-pgpool.${namespace}.svc.cluster.local
PG_PORT=5432
PG_DATABASE=yatai
PG_USER=postgres
PG_SSLMODE=disable

echo "🧪 testing PostgreSQL connection..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2> /dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace ${namespace} \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "SELECT 1"

echo "✅ PostgreSQL connection is successful"

echo "🧐 checking if PostgreSQL database ${PG_DATABASE} exists..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2> /dev/null || true
if ! kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace ${namespace} \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d ${PG_DATABASE} -c "SELECT 1" > /dev/null 2>&1; then

  echo "🥹 PostgreSQL database ${PG_DATABASE} does not exist"
  echo "🤖 creating PostgreSQL database ${PG_DATABASE}..."
  kubectl -n ${namespace} delete pod postgresql-ha-client 2> /dev/null || true

  kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
      --namespace ${namespace} \
      --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
      --env="PGPASSWORD=$PG_PASSWORD" \
      --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "CREATE DATABASE $PG_DATABASE"

  echo "✅ PostgreSQL database ${PG_DATABASE} is created"
else
  echo "🤩 PostgreSQL database ${PG_DATABASE} already exists"
fi

echo "🧪 testing PostgreSQL environment variables..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2> /dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace ${namespace} \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -c "select 1"

echo "✅ PostgreSQL environment variables are correct"

helm repo add minio https://operator.min.io/
helm repo update minio

echo "🤖 installing minio-operator..."
helm upgrade --install minio-operator minio/minio-operator -n ${namespace} --set tenants=null

echo "⏳ waiting for minio-operator to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=minio-operator
echo "✅ minio-operator is ready"

minio_secret_name=yatai-minio

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

echo "🤖 creating MinIO Tenant..."
cat <<EOF | kubectl apply -f -
apiVersion: minio.min.io/v2
kind: Tenant
metadata:
  labels:
    app: yatai-minio
  name: yatai-minio
  namespace: ${namespace}
spec:
  credsSecret:
    name: ${minio_secret_name}
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
# this retry logic is to avoid kubectl wait errors due to minio tenant resources not being created
for i in $(seq 1 10); do
  kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app=yatai-minio && break || sleep 5
done
echo "✅ minio tenant is ready"

S3_ENDPOINT=minio.${namespace}.svc.cluster.local
S3_REGION=foo
S3_BUCKET_NAME=yatai
S3_SECURE=false
S3_ACCESS_KEY=$(kubectl -n ${namespace} get secret ${minio_secret_name} -o jsonpath='{.data.accesskey}' | base64 -d)
S3_SECRET_KEY=$(kubectl -n ${namespace} get secret ${minio_secret_name} -o jsonpath='{.data.secretkey}' | base64 -d)

# check if S3_ACCESS_KEY is empty
if [ -z "$S3_ACCESS_KEY" ]; then
  echo "🥹 S3_ACCESS_KEY is empty" >&2
  exit 1
fi

echo "🧪 testing MinIO connection..."
for i in $(seq 1 10); do
  kubectl -n ${namespace} delete pod s3-client 2> /dev/null || true

  kubectl run s3-client --rm --tty -i --restart='Never' \
      --namespace ${namespace} \
      --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
      --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
      --image quay.io/bentoml/s3-client:0.0.1 \
      --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets 2>/dev/null" && break || sleep 5
done
echo "✅ MinIO connection is successful"

helm_repo_name=bentoml
helm_repo_url=https://bentoml.github.io/helm-charts

# check if DEVEL_HELM_REPO is true
if [ "${DEVEL_HELM_REPO}" = "true" ]; then
  helm_repo_name=bentoml-devel
  helm_repo_url=https://bentoml.github.io/helm-charts-devel
fi

helm repo remove ${helm_repo_name} 2> /dev/null || true
helm repo add ${helm_repo_name} ${helm_repo_url}
helm repo update ${helm_repo_name}
echo "🤖 installing yatai..."
helm upgrade --install yatai ${helm_repo_name}/yatai -n ${namespace} \
    --set postgresql.host=$PG_HOST \
    --set postgresql.port=$PG_PORT \
    --set postgresql.user=$PG_USER \
    --set postgresql.database=$PG_DATABASE \
    --set postgresql.password=$PG_PASSWORD \
    --set postgresql.sslmode=$PG_SSLMODE \
    --set s3.endpoint=$S3_ENDPOINT \
    --set s3.region=$S3_REGION \
    --set s3.bucketName=$S3_BUCKET_NAME \
    --set s3.secure=$S3_SECURE \
    --set s3.accessKey=$S3_ACCESS_KEY \
    --set s3.secretKey=$S3_SECRET_KEY \
    --devel=$DEVEL

echo "⏳ waiting for yatai to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=yatai
echo "✅ yatai is ready"
helm get notes yatai -n ${namespace}
