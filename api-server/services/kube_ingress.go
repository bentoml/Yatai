package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bentoml/yatai/common/utils"

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

func (s *kubeIngressService) ToKubeIngresses(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (ingresses []*v1.Ingress, err error) {
	kubeName, err := DeploymentTargetService.GetKubeName(ctx, deploymentTarget)
	if err != nil {
		return
	}

	bentoVersion, err := BentoVersionService.GetAssociatedBentoVersion(ctx, deploymentTarget)
	if err != nil {
		return
	}

	bento, err := BentoService.GetAssociatedBento(ctx, bentoVersion)
	if err != nil {
		return
	}

	internalHost, err := DeploymentTargetService.GenerateIngressHost(ctx, deploymentTarget)
	if err != nil {
		return
	}

	annotations, err := DeploymentTargetService.GetKubeAnnotations(ctx, deploymentTarget)
	if err != nil {
		return
	}

	annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf(`
more_set_headers "X-Powered-By: Yatai";
more_set_headers "X-Yatai-Bento: %s";
more_set_headers "X-Yatai-Bento-Version: %s";
`, bento.Name, bentoVersion.Version)
	if deploymentTarget.Type == modelschemas.DeploymentTargetTypeCanary && deploymentTarget.CanaryRules != nil {
		annotations["nginx.ingress.kubernetes.io/canary"] = "true"
		for _, rule := range *deploymentTarget.CanaryRules {
			// nolint: gocritic
			if rule.Type == modelschemas.DeploymentTargetCanaryRuleTypeWeight && rule.Weight != nil {
				annotations["nginx.ingress.kubernetes.io/canary-weight"] = strconv.Itoa(int(*rule.Weight))
			} else if rule.Type == modelschemas.DeploymentTargetCanaryRuleTypeHeader && rule.Header != nil {
				annotations["nginx.ingress.kubernetes.io/canary-by-header"] = *rule.Header
				if rule.HeaderValue != nil {
					annotations["nginx.ingress.kubernetes.io/canary-by-header-value"] = *rule.HeaderValue
				}
			} else if rule.Type == modelschemas.DeploymentTargetCanaryRuleTypeCookie && rule.Cookie != nil {
				annotations["nginx.ingress.kubernetes.io/canary-by-cookie"] = *rule.Cookie
			}
		}
	}

	annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "false"

	labels, err := DeploymentTargetService.GetKubeLabels(ctx, deploymentTarget)
	if err != nil {
		return
	}

	pathType := v1.PathTypeImplementationSpecific

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return nil, errors.Wrap(err, "get deployment")
	}

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	interIng := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       kubeNs,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: v1.IngressSpec{
			IngressClassName: utils.StringPtr(consts.KubeIngressClassName),
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

func (s *kubeIngressService) DeployDeploymentTargetAsKubeIngresses(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) error {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}

	ingressesCli, err := DeploymentService.GetKubeIngressesCli(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get kube ingresses cli")
	}

	kubeIngresses, err := s.ToKubeIngresses(ctx, deploymentTarget, deployOption)
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
