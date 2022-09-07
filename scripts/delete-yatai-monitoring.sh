#!/usr/bin/env bash

CURRETN_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: The Prometheus, Grafana and Alertmanager under the yatai-monitoring namespace will be removed.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-monitoring\033[00m \033[01;31mnamespace will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m$CURRETN_CONTEXT\033[00m"

while true; do
  echo -e -n "Are you sure to delete yatai-monitoring in cluster \033[00m\033[01;32m${CURRETN_CONTEXT}\033[00m? [y/n] "
  read yn
  case $yn in
    [Yy]* ) break;;
    [Nn]* ) exit;;
    * ) echo "Please answer yes or no.";;
  esac
done

echo "Uninstalling helm releases in yatai-monitoring namespace from cluster.."
set -x
helm list -n yatai-monitoring | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-monitoring uninstall {}
set +x

echo "Removing additional yatai-monitoring related namespaces and resources.."
set -x
kubectl delete namespace yatai-monitoring
set +x

echo "Done"
