#!/usr/bin/env bash

CURRENT_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: The MinIO tenant, Loki and Promtail under the yatai-logging namespace will be removed.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-logging\033[00m \033[01;31mnamespace will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m$CURRENT_CONTEXT\033[00m"

while true; do
  echo -e -n "Are you sure to delete yatai-logging in cluster \033[00m\033[01;32m${CURRENT_CONTEXT}\033[00m? [y/n] "
  read yn
  case $yn in
    [Yy]* ) break;;
    [Nn]* ) exit;;
    * ) echo "Please answer yes or no.";;
  esac
done

echo "Uninstalling helm releases in yatai-logging namespace from cluster.."
set -x
helm list -n yatai-logging | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-logging uninstall {}
set +x

echo "Removing additional yatai-logging related namespaces and resources.."
set -x
kubectl delete namespace yatai-logging
set +x

echo "Done"
