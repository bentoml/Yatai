package services

import (
	"context"
	"fmt"
	"strconv"

	v1 "k8s.io/api/networking/v1"

	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai/api-server/models"
)

type kubeIngressService struct{}

var KubeIngressService = kubeIngressService{}

func (s *kubeIngressService) ToKubeIngresses(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) (ingresses []*v1.Ingress, err error) {
	kubeName, err := DeploymentSnapshotService.GetKubeName(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	bentoVersion, err := BentoVersionService.GetAssociatedBentoVersion(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return
	}

	internalHost, err := DeploymentSnapshotService.GetIngressHost(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	annotations, err := DeploymentSnapshotService.GetKubeAnnotations(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf(`
more_set_headers "X-Powered-By: Yatai";
more_set_headers "X-Yatai-Bento: %s";
more_set_headers "X-Yatai-Bento-Version: %s";
`, bento.Name, bentoVersion.Version)
	if deploymentSnapshot.Type == modelschemas.DeploymentSnapshotTypeCanary && deploymentSnapshot.CanaryRules != nil {
		annotations["nginx.ingress.kubernetes.io/canary"] = "true"
		for _, rule := range *deploymentSnapshot.CanaryRules {
			// nolint: gocritic
			if rule.Type == modelschemas.DeploymentSnapshotCanaryRuleTypeWeight && rule.Weight != nil {
				annotations["nginx.ingress.kubernetes.io/canary-weight"] = strconv.Itoa(int(*rule.Weight))
			} else if rule.Type == modelschemas.DeploymentSnapshotCanaryRuleTypeHeader && rule.Header != nil {
				annotations["nginx.ingress.kubernetes.io/canary-by-header"] = *rule.Header
				if rule.HeaderValue != nil {
					annotations["nginx.ingress.kubernetes.io/canary-by-header-value"] = *rule.HeaderValue
				}
			} else if rule.Type == modelschemas.DeploymentSnapshotCanaryRuleTypeCookie && rule.Cookie != nil {
				annotations["nginx.ingress.kubernetes.io/canary-by-cookie"] = *rule.Cookie
			}
		}
	}

	annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "false"

	labels, err := DeploymentSnapshotService.GetKubeLabels(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	pathType := v1.PathTypeImplementationSpecific

	interIng := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       consts.KubeNamespaceYataiDeployment,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: internalHost,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: kubeName,
											Port: v1.ServiceBackendPort{
												Number: consts.BentoServicePort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ings := []*v1.Ingress{interIng}

	return ings, nil
}

func (s *kubeIngressService) DeployDeploymentSnapshotAsKubeIngresses(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) error {
	kubeCli, _, err := DeploymentSnapshotService.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return err
	}
	ingressesCli := kubeCli.NetworkingV1().Ingresses(consts.KubeNamespaceYataiDeployment)
	kubeIngresses, err := s.ToKubeIngresses(ctx, deploymentSnapshot, deployOption)
	if err != nil {
		return err
	}
	for _, kubeIng := range kubeIngresses {
		logrus.Infof("get k8s ingress %s ...", kubeIng.Name)
		_, err = ingressesCli.Get(ctx, kubeIng.Name, metav1.GetOptions{})
		notFound := apierrors.IsNotFound(err)
		if !notFound && err != nil {
			return errors.Wrapf(err, "get k8s ingress %s", kubeIng.Name)
		}
		if notFound {
			logrus.Infof("create k8s ingress %s ...", kubeIng.Name)
			_, err = ingressesCli.Create(ctx, kubeIng, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "create k8s ingress %s", kubeIng.Name)
			}
		} else {
			logrus.Infof("update k8s ingress %s ...", kubeIng.Name)
			_, err = ingressesCli.Update(ctx, kubeIng, metav1.UpdateOptions{})
			if err != nil {
				return errors.Wrapf(err, "update k8s ingress %s", kubeIng.Name)
			}
		}
	}
	return nil
}
