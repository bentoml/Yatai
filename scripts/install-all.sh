set -e

DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/quick-install-yatai.sh")
DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/v1.0.0/scripts/quick-install-yatai-deployment.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/quick-setup-yatai-monitoring.sh")
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/quick-setup-yatai-logging.sh")
