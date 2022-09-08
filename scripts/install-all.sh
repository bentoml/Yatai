#!/usr/bin/env bash

set -e

DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-install-yatai.sh")
DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/main/scripts/quick-install-yatai-deployment.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-setup-yatai-monitoring.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-setup-yatai-logging.sh")

helm get notes yatai -n yatai-system
