#!/usr/bin/env bash

set -e

DEVEL=${DEVEL:-false}
DEVEL_HELM_REPO=${DEVEL_HELM_REPO:-false}

function randstr() {
  LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 20
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
if [ "${KUBE_VERSION:1:2}" -lt 20 ]; then
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
  postgresql_password=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.password}" | base64 -d)
  repmgr_password=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.repmgr-password}" | base64 -d)
fi

if ! kubectl -n ${namespace} get secret postgresql-ha-pgpool >/dev/null 2>&1; then
  pgpool_admin_password=$(randstr)
else
  pgpool_admin_password=$(kubectl -n ${namespace} get secret postgresql-ha-pgpool -o jsonpath="{.data.admin-password}" | base64 -d)
fi

# get installed postgresql-ha version
postgresql_ha_version=$(helm list -n ${namespace} -o json | $jq -r '.[] | select(.name == "postgresql-ha") | .chart' | cut -d'/' -f2 | cut -d'-' -f3)
# if postgresql-ha already installed but version is less than 10.0.6, get postgresql_password with the old way
if [[ -n "${postgresql_ha_version}" && "${postgresql_ha_version}" < "10.0.6" ]]; then
  echo "🤖 postgresql-ha already installed, but it's version ${postgresql_ha_version} is less than 10.0.6, getting postgresql password with the old way"
  postgresql_password=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.postgresql-password}" | base64 -d)
fi

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update bitnami
echo "🤖 installing PostgreSQL..."
helm upgrade --install postgresql-ha bitnami/postgresql-ha -n ${namespace} \
  --set postgresql.password="${postgresql_password}" \
  --set postgresql.repmgrPassword="${repmgr_password}" \
  --set pgpool.adminPassword="${pgpool_admin_password}" \
  --version 10.0.6

echo "⏳ waiting for PostgreSQL to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=postgresql-ha
echo "✅ PostgreSQL is ready"

PG_PASSWORD=$(kubectl -n ${namespace} get secret postgresql-ha-postgresql -o jsonpath="{.data.password}" | base64 -d)
PG_HOST=postgresql-ha-pgpool.${namespace}.svc.cluster.local
PG_PORT=5432
PG_DATABASE=yatai
PG_USER=postgres
PG_SSLMODE=disable

echo "🧪 testing PostgreSQL connection..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2>/dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
  --namespace ${namespace} \
  --image postgres:14.4-alpine \
  --env="PGPASSWORD=$PG_PASSWORD" \
  --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "SELECT 1"

echo "✅ PostgreSQL connection is successful"

echo "🧐 checking if PostgreSQL database ${PG_DATABASE} exists..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2>/dev/null || true
if ! kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
  --namespace ${namespace} \
  --image postgres:14.4-alpine \
  --env="PGPASSWORD=$PG_PASSWORD" \
  --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d ${PG_DATABASE} -c "SELECT 1" >/dev/null 2>&1; then

  echo "🥹 PostgreSQL database ${PG_DATABASE} does not exist"
  echo "🤖 creating PostgreSQL database ${PG_DATABASE}..."
  kubectl -n ${namespace} delete pod postgresql-ha-client 2>/dev/null || true

  kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace ${namespace} \
    --image postgres:14.4-alpine \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "CREATE DATABASE $PG_DATABASE"

  echo "✅ PostgreSQL database ${PG_DATABASE} is created"
else
  echo "🤩 PostgreSQL database ${PG_DATABASE} already exists"
fi

echo "🧪 testing PostgreSQL environment variables..."
kubectl -n ${namespace} delete pod postgresql-ha-client 2>/dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
  --namespace ${namespace} \
  --image postgres:14.4-alpine \
  --env="PGPASSWORD=$PG_PASSWORD" \
  --command -- psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -c "select 1"

echo "✅ PostgreSQL environment variables are correct"

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

minio_secret_name=yatai-minio

# check if logging minio secret not exists
echo "🧐 checking if secret ${minio_secret_name} exists..."
if ! kubectl get secret ${minio_secret_name} -n ${namespace} >/dev/null 2>&1; then
  echo "🥹 secret ${minio_secret_name} not found"
  echo "🤖 creating secret ${minio_secret_name}"
  kubectl create secret generic ${minio_secret_name} \
    --from-literal=accesskey="$(randstr)" \
    --from-literal=secretkey="$(randstr)" \
    -n ${namespace}
  echo "✅ created secret ${minio_secret_name}"
else
  echo "🤩 secret ${minio_secret_name} already exists"
fi

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
helm upgrade --install yatai-minio-tenant minio/tenant \
  -n ${namespace} \
  --set secrets.accessKey=${S3_ACCESS_KEY} \
  --set secrets.secretKey=${S3_SECRET_KEY} \
  --set tenant.name=yatai-minio \
  --set tenant.certificate.requestAutoCert=false


echo "⏳ waiting for minio tenant to be ready..."
# this retry logic is to avoid kubectl wait errors due to minio tenant resources not being created
for i in $(seq 1 10); do
  if kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l v1.min.io/tenant=yatai-minio; then
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
  kubectl -n ${namespace} delete pod s3-client 2>/dev/null || true

  if kubectl run s3-client --rm --tty -i --restart='Never' \
    --namespace ${namespace} \
    --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
    --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
    --image quay.io/bentoml/s3-client:0.0.1 \
    --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets 2>/dev/null"; then
      echo "✅ MinIO connection is successful"
      break
    else
      if [ $i -eq 10 ]; then
        echo "😱 MinIO connection is not successful"
        exit 1
      fi
      echo "😱 MinIO connection is not successful, retrying..."
      sleep 5
      continue
  fi
done

helm_repo_name=bentoml
helm_repo_url=https://bentoml.github.io/helm-charts

# check if DEVEL_HELM_REPO is true
if [ "${DEVEL_HELM_REPO}" = "true" ]; then
  helm_repo_name=bentoml-devel
  helm_repo_url=https://bentoml.github.io/helm-charts-devel
fi

helm repo remove ${helm_repo_name} 2>/dev/null || true
helm repo add ${helm_repo_name} ${helm_repo_url}
helm repo update ${helm_repo_name}

# if $VERSION is not set, use the latest version
if [ -z "$VERSION" ]; then
  VERSION=$(helm search repo ${helm_repo_name} --devel="$DEVEL" -l | grep "${helm_repo_name}/yatai " | awk '{print $2}' | head -n 1)
fi

echo "🤖 installing yatai ${VERSION} from helm repo ${helm_repo_name}..."
helm upgrade --install yatai ${helm_repo_name}/yatai -n ${namespace} \
  --set postgresql.host="$PG_HOST" \
  --set postgresql.port="$PG_PORT" \
  --set postgresql.user="$PG_USER" \
  --set postgresql.database="$PG_DATABASE" \
  --set postgresql.password="$PG_PASSWORD" \
  --set postgresql.sslmode="$PG_SSLMODE" \
  --set s3.endpoint="$S3_ENDPOINT" \
  --set s3.region="$S3_REGION" \
  --set s3.bucketName="$S3_BUCKET_NAME" \
  --set s3.secure="$S3_SECURE" \
  --set s3.accessKey="$S3_ACCESS_KEY" \
  --set s3.secretKey="$S3_SECRET_KEY" \
  --devel="$DEVEL" \
  --version "$VERSION"

echo "⏳ waiting for yatai to be ready..."
kubectl -n ${namespace} wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=yatai
echo "✅ yatai is ready"
helm get notes yatai -n ${namespace}
