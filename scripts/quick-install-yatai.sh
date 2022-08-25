#!/bin/bash

set -e

DEVEL=${DEVEL:-false}

# check if jq command exists
if ! command -v jq >/dev/null 2>&1; then
  echo "😱 jq command not found, please install it first!" >&2
  exit 1
fi

# check if kubectl command exists
if ! command -v kubectl >/dev/null 2>&1; then
  echo "😱 kubectl command is not found, please install it first!" >&2
  exit 1
fi

KUBE_VERSION=$(kubectl version --output=json | jq '.serverVersion.minor')
if [ ${KUBE_VERSION} -lt 20 ]; then
  echo "😱 install requires at least Kubernetes 1.20" >&2
  exit 1
fi

# check if helm command exists
if ! command -v helm >/dev/null 2>&1; then
  echo "😱 helm command is not found, please install it first!" >&2
  exit 1
fi

kubectl create namespace yatai-system

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update bitnami
echo "🤖 installing PostgreSQL..."
helm install postgresql-ha bitnami/postgresql-ha -n yatai-system

echo "⏳ waiting for PostgreSQL to be ready..."
kubectl -n yatai-system wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=postgresql-ha
echo "✅ PostgreSQL is ready"

PG_PASSWORD=$(kubectl get secret --namespace yatai-system postgresql-ha-postgresql -o jsonpath="{.data.postgresql-password}" | base64 -d)
PG_HOST=postgresql-ha-pgpool.yatai-system.svc.cluster.local
PG_PORT=5432
PG_DATABASE=yatai
PG_USER=postgres
PG_SSLMODE=disable

echo "🧪 testing PostgreSQL connection..."
kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace yatai-system \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "select 1"

echo "✅ PostgreSQL connection is successful"

echo "🤖 creating PostgreSQL database ${PG_DATABASE}..."
kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace yatai-system \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "create database $PG_DATABASE"

echo "✅ PostgreSQL database ${PG_DATABASE} is created"

echo "🧪 testing PostgreSQL environment variables..."
kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true

kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
    --namespace yatai-system \
    --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
    --env="PGPASSWORD=$PG_PASSWORD" \
    --command -- psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -c "select 1"

echo "✅ PostgreSQL environment variables are correct"

helm repo add minio https://operator.min.io/
helm repo update minio

S3_ACCESS_KEY=$(echo $RANDOM | md5sum | head -c 20; echo -n)
S3_SECRET_KEY=$(echo $RANDOM | md5sum | head -c 20; echo -n)

cat <<EOF > /tmp/yatai-minio-values.yaml
tenants:
- image:
    pullPolicy: IfNotPresent
    repository: quay.io/bentoml/minio-minio
    tag: RELEASE.2021-10-06T23-36-31Z
  metrics:
    enabled: false
    port: 9000
  mountPath: /export
  name: yatai-minio
  namespace: yatai-system
  pools:
  - servers: 4
    size: 20Gi
    volumesPerServer: 4
  secrets:
    accessKey: $S3_ACCESS_KEY
    enabled: true
    name: yatai-minio-secret
    secretKey: $S3_SECRET_KEY
  subPath: /data
EOF

echo "🤖 installing MinIO..."
helm install minio-operator minio/minio-operator -n yatai-system -f /tmp/yatai-minio-values.yaml

echo "⏳ waiting for minio-operator to be ready..."
kubectl -n yatai-system wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=minio-operator
echo "✅ minio-operator is ready"

echo "⏳ waiting for minio tenant to be ready..."
for i in $(seq 1 10); do
  kubectl -n yatai-system wait --for=condition=ready --timeout=600s pod -l app=minio && break || sleep 5
done
echo "✅ minio tenant is ready"

S3_ENDPOINT=minio.yatai-system.svc.cluster.local
S3_REGION=foo
S3_BUCKET_NAME=yatai
S3_SECURE=false
S3_ACCESS_KEY=$(kubectl -n yatai-system get secret yatai-minio-secret -o jsonpath='{.data.accesskey}' | base64 -d)
S3_SECRET_KEY=$(kubectl -n yatai-system get secret yatai-minio-secret -o jsonpath='{.data.secretkey}' | base64 -d)

echo "🧪 testing MinIO connection..."
for i in $(seq 1 10); do
  kubectl -n yatai-system delete pod s3-client 2> /dev/null || true

  kubectl run s3-client --rm --tty -i --restart='Never' \
      --namespace yatai-system \
      --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
      --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
      --image quay.io/bentoml/s3-client:0.0.1 \
      --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets 2>/dev/null && echo successfully || echo failed" && break || sleep 5
done
echo "✅ MinIO connection is successful"

helm repo add bentoml https://bentoml.github.io/charts
helm repo update bentoml
echo "🤖 installing yatai..."
helm install yatai bentoml/yatai -n yatai-system \
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
kubectl -n yatai-system wait --for=condition=ready --timeout=600s pod -l app.kubernetes.io/name=yatai
echo "✅ yatai is ready"
helm get notes yatai -n yatai-system
