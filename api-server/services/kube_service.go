package services

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type kubeServiceService struct{}

var KubeServiceService = kubeServiceService{}

func (s *kubeServiceService) DeploymentTargetToKubeService(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeService *apiv1.Service, err error) {
	kubeName, err := DeploymentTargetService.GetKubeName(ctx, deploymentTarget)
	if err != nil {
		return
	}

	targetPort := consts.BentoServicePort
	if deploymentTarget.Config != nil && deploymentTarget.Config.Envs != nil {
		for _, env := range *deploymentTarget.Config.Envs {
			if env.Key == consts.BentoServicePortEnvKey {
				port_, err := strconv.Atoi(env.Value)
				if err != nil {
					return nil, errors.Wrapf(err, "convert port %s to int", env.Value)
				}
				targetPort = port_
				break
			}
		}
	}

	spec := apiv1.ServiceSpec{
		Selector: map[string]string{
			consts.KubeLabelYataiSelector: kubeName,
		},
		Ports: []apiv1.ServicePort{
			{
				Name:       "http-default",
				Port:       consts.BentoServicePort,
				TargetPort: intstr.FromInt(targetPort),
				Protocol:   apiv1.ProtocolTCP,
			},
		},
	}

	labels, err := DeploymentTargetService.GetKubeLabels(ctx, deploymentTarget)
	if err != nil {
		return
	}

	annotations, err := DeploymentTargetService.GetKubeAnnotations(ctx, deploymentTarget)
	if err != nil {
		return
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "get deployment")
		return
	}

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	kubeService = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       kubeNs,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: spec,
	}

	return
}

func (s *kubeServiceService) DeployKubeService(ctx context.Context, deploymentTarget *models.DeploymentTarget, kubeService *apiv1.Service) error {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}

	servicesCli, err := DeploymentService.GetKubeServicesCli(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get kube services cli")
	}

	logrus.Infof("get k8s service %s ...", kubeService.Name)
	oldSvc, err := servicesCli.Get(ctx, kubeService.Name, metav1.GetOptions{})
	notFound := apierrors.IsNotFound(err)
	if !notFound && err != nil {
		return errors.Wrapf(err, "get k8s service %s", kubeService.Name)
	}
	if notFound {
		logrus.Infof("create k8s service %s ...", kubeService.Name)
		_, err = servicesCli.Create(ctx, kubeService, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "create k8s service %s", kubeService.Name)
		}
	} else {
		if (kubeService.Spec.ClusterIP == consts.NoneStr || oldSvc.Spec.ClusterIP == consts.NoneStr) && kubeService.Spec.ClusterIP != oldSvc.Spec.ClusterIP {
			logrus.Infof("delete old k8s service %s ...", oldSvc.Name)
			err = s.DeleteKubeService(ctx, deploymentTarget, oldSvc.Name)
			if err != nil {
				return errors.Wrapf(err, "delete old k8s service %s", oldSvc.Name)
			}
			logrus.Infof("create k8s service %s ...", kubeService.Name)
			_, err = servicesCli.Create(ctx, kubeService, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "create k8s service %s", kubeService.Name)
			}
		} else {
			logrus.Infof("update k8s service %s ...", kubeService.Name)
			kubeService.ObjectMeta.ResourceVersion = oldSvc.ObjectMeta.ResourceVersion
			if kubeService.Spec.Type != apiv1.ServiceTypeExternalName {
				// Service is invalid: spec.clusterIP: Invalid value: "": field is immutable
				kubeService.Spec.ClusterIP = oldSvc.Spec.ClusterIP
			}
			if kubeService.Spec.Type == apiv1.ServiceTypeNodePort {
				for i, port := range kubeService.Spec.Ports {
					for _, oldPort := range oldSvc.Spec.Ports {
						if oldPort.TargetPort == port.TargetPort {
							port.NodePort = oldPort.NodePort
							kubeService.Spec.Ports[i] = port
							break
						}
					}
				}
			}
			_, err = servicesCli.Update(ctx, kubeService, metav1.UpdateOptions{})
			if err != nil {
				return errors.Wrapf(err, "update k8s service %s", kubeService.Name)
			}
		}
	}
	return err
}

func (s *kubeServiceService) DeleteKubeService(ctx context.Context, deploymentTarget *models.DeploymentTarget, kubeServiceName string) error {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}

	servicesCli, err := DeploymentService.GetKubeServicesCli(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get kube services cli")
	}
	logrus.Infof("delete k8s service %s ...", kubeServiceName)
	return servicesCli.Delete(ctx, kubeServiceName, metav1.DeleteOptions{})
}

func (s *kubeServiceService) DeployDeploymentTargetAsKubeService(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) error {
	kubeService, err := s.DeploymentTargetToKubeService(ctx, deploymentTarget, deployOption)
	if err != nil {
		return errors.Wrap(err, "to k8s service")
	}

	err = s.DeployKubeService(ctx, deploymentTarget, kubeService)
	return err
}
