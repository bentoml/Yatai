package services

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai/api-server/models"
)

type kubeNamespaceService struct{}

var KubeNamespaceService = kubeNamespaceService{}

func (s *kubeNamespaceService) MakeSureNamespace(ctx context.Context, cluster *models.Cluster, namespace string) (kubeNs *apiv1.Namespace, err error) {
	kubeCli, _, err := ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return
	}

	nsCli := kubeCli.CoreV1().Namespaces()
	kubeNs, err = nsCli.Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		kubeNs = &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		}}
		err = nil
		_, err_ := nsCli.Create(ctx, kubeNs, metav1.CreateOptions{})
		if err_ != nil {
			kubeNs, err = nsCli.Get(ctx, namespace, metav1.GetOptions{})
			if err == nil {
				return // nolint: nilerr
			}
			err = err_
			return
		}
	} else if err != nil {
		return
	}

	return
}
