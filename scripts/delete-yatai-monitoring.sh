#!/usr/bin/env bash

CURRETN_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: The Prometheus, Grafana and Alertmanager under the yatai-monitoring namespace will be removed.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-monitoring\033[00m \033[01;31mnamespace will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m$CURRETN_CONTEXT\033[00m"
read -p "Are you sure to delete yatai-monitoring in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi
read -p "(Double check) Are you sure to delete yatai-monitoring in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi

echo "Uninstalling helm releases in yatai-monitoring namespace from cluster.."
set -x
helm list -n yatai-monitoring | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-monitoring uninstall {}
set +x

echo "Removing additional yatai-monitoring related namespaces and resources.."
set -x
kubectl delete namespace yatai-monitoring
set +x

echo "Done"
