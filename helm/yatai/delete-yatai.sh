#!/usr/bin/env bash

CURRETN_CONTEXT=$(kubectl config current-context)
echo -e "\033[01;31mWarning: this will permanently delete all Yatai resources, existing model deployments, and in-cluster minio, postgresql DB data. Note that external DB and blob storage will not be deleted.\033[00m"
echo -e "\033[01;31mWarning: this also means that all resources under the \033[00m\033[01;32myatai\033[00m, \033[00m\033[01;32myatai-system\033[00m, \033[00m\033[01;32myatai-operators\033[00m, and \033[00m\033[01;32myatai-components\033[00m \033[01;31mnamespaces will be permanently deleted.\033[00m"
echo -e "\033[01;31mCurrent kubernetes context: \033[00m\033[01;32m$CURRETN_CONTEXT\033[00m"
read -p "Are you sure to delete Yatai in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
read -p "(Double check) Are you sure to delete Yatai in cluster '$CURRETN_CONTEXT'? [y/n] " -n 1 -r
echo # move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi

echo "Uninstalling yatai helm chart from cluster.."
set -x
helm list -n yatai-system | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-system uninstall {}
set +x

echo "Uninstalling yatai-deployment helm chart from cluster.."
set -x
helm list -n yatai-deployment | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-deployment uninstall {}
set +x

echo "Deleting all crd resources and BentoDeployment.."
set -x
kubectl delete crd bentodeployments.serving.yatai.ai
kubectl delete crd deployments.component.yatai.ai
set +x

echo "Removing additional yatai related namespaces and resources.."
set -x
kubectl delete namespace yatai
kubectl delete namespace yatai-builders
kubectl delete namespace yatai-deployment
kubectl delete namespace yatai-system
set +x

echo "Done"
