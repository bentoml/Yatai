#!/usr/bin/env bash

CURRENT_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: this will permanently delete all Yatai resources, in-cluster minio, postgresql DB data. Note that external DB and blob storage will not be deleted.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai-system\033[00m \033[01;31mnamespace will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m${CURRENT_CONTEXT}\033[00m"

while true; do
  echo -e -n "Are you sure to delete Yatai in cluster \033[00m\033[01;32m${CURRENT_CONTEXT}\033[00m? [y/n] "
  read yn
  case $yn in
    [Yy]* ) break;;
    [Nn]* ) exit;;
    * ) echo "Please answer yes or no.";;
  esac
done

echo "Uninstalling yatai helm chart from cluster.."
set -x
helm list -n yatai-system | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-system uninstall {}
set +x

echo "Removing additional yatai related namespaces and resources.."
set -x
kubectl delete namespace yatai-system
set +x

echo "Done"
