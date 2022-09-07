#!/usr/bin/env bash

set -e

CURRETN_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: this will permanently delete all Yatai resources, in-cluster minio, postgresql DB data. Note that external DB and blob storage will not be deleted.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-system\033[00m, \033[00m\033[01;32myatai-deployment\033[00m, \033[00m\033[01;32myatai-builders\033[00m, \033[00m\033[01;32myatai\033[00m, \033[00m\033[01;32myatai-monitoring\033[00m, \033[00m\033[01;32myatai-logging\033[00m \033[01;31mnamespaces will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m${CURRETN_CONTEXT}\033[00m"

while true; do
  echo -e -n "Are you sure to delete all components of Yatai in cluster \033[00m\033[01;32m${CURRETN_CONTEXT}\033[00m? [y/n] "
  read yn
  case $yn in
    [Yy]* ) break;;
    [Nn]* ) exit;;
    * ) echo "Please answer yes or no.";;
  esac
done

yes | bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai-logging.sh")
yes | bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai-monitoring.sh")
yes | bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/v1.0.0/scripts/delete-yatai-deployment.sh")
yes | bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai.sh")
