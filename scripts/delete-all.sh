#!/usr/bin/env bash

set -e

CURRETN_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: this will permanently delete all Yatai resources, in-cluster minio, postgresql DB data. Note that external DB and blob storage will not be deleted.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-system\033[00m, \033[00m\033[01;32myatai-deployment\033[00m, \033[00m\033[01;32myatai-builders\033[00m, \033[00m\033[01;32myatai\033[00m, \033[00m\033[01;32myatai-monitoring\033[00m, \033[00m\033[01;32myatai-logging\033[00m \033[01;31mnamespaces will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m$CURRETN_CONTEXT\033[00m"
read -p "Are you sure to delete Yatai in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi
read -p "(Double check) Are you sure to delete Yatai in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi
curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai-logging.sh" | bash
curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai-monitoring.sh" | bash
curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/v1.0.0/scripts/delete-yatai-deployment.sh" | bash
curl -s "https://raw.githubusercontent.com/bentoml/yatai/v1.0.0/scripts/delete-yatai.sh" | bash
