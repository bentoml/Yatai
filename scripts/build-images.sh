set -x

DOCKER_REPOSITORY=quay.io/bentoml

images=(
registry:2.7.1
minio/operator:v4.3.5
minio/console:v0.12.3
busybox:1.31.1
grafana/grafana-image-renderer
grafana/grafana:8.2.0
curlimages/curl:7.73.0
nginxinc/nginx-unprivileged:1.19-alpine
grafana/loki
grafana/loki:2.4.1
busybox:1.33
grafana/promtail
grafana/promtail:2.4.1
grafana/promtail:2.1.0
yetone/grafana:8.2.0
grafana/promtail:2.4.1
busybox
minio/minio:RELEASE.2021-10-06T23-36-31Z
bitnami/postgresql-repmgr:11.14.0-debian-10-r12
)

for image in "${images[@]}"; do
    docker pull ${image}
    newImage=${DOCKER_REPOSITORY}/${image/\//-}
    docker tag ${image} ${newImage}
    docker push ${newImage}
done
