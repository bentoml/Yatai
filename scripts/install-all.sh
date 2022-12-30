#!/usr/bin/env bash

set -e

bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-install-yatai.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/v1.0.x/scripts/quick-install-yatai-deployment.sh")
kubectl -n yatai-deployment rollout restart deploy/yatai-deployment || true
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-setup-yatai-monitoring.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-setup-yatai-logging.sh")

helm get notes yatai -n yatai-system
