package services

import (
	"context"

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

func (s *kubeServiceService) DeploymentSnapshotToKubeService(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) (kubeService *apiv1.Service, err error) {
	kubeName, err := DeploymentSnapshotService.GetKubeName(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	spec := apiv1.ServiceSpec{
		Selector: map[string]string{
			consts.KubeLabelYataiSelector: kubeName,
		},
		Ports: []apiv1.ServicePort{
			{
				Name:       "http-default",
				Port:       consts.BentoServicePort,
				TargetPort: intstr.FromInt(consts.BentoServicePort),
				Protocol:   apiv1.ProtocolTCP,
			},
		},
	}

	labels, err := DeploymentSnapshotService.GetKubeLabels(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	annotations, err := DeploymentSnapshotService.GetKubeAnnotations(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	kubeService = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       consts.KubeNamespaceYataiDeployment,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: spec,
	}

	return
}

func (s *kubeServiceService) DeployKubeService(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, kubeService *apiv1.Service) error {
	kubeCli, _, err := DeploymentSnapshotService.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return err
	}
	servicesCli := kubeCli.CoreV1().Services(consts.KubeNamespaceYataiDeployment)
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
			err = s.DeleteKubeService(ctx, deploymentSnapshot, oldSvc.Name)
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

func (s *kubeServiceService) DeleteKubeService(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, kubeServiceName string) error {
	kubeCli, _, err := DeploymentSnapshotService.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return err
	}
	servicesCli := kubeCli.CoreV1().Services(consts.KubeNamespaceYataiDeployment)
	logrus.Infof("delete k8s service %s ...", kubeServiceName)
	return servicesCli.Delete(ctx, kubeServiceName, metav1.DeleteOptions{})
}

func (s *kubeServiceService) DeployDeploymentSnapshotAsKubeService(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) error {
	kubeService, err := s.DeploymentSnapshotToKubeService(ctx, deploymentSnapshot, deployOption)
	if err != nil {
		return errors.Wrap(err, "to k8s service")
	}

	err = s.DeployKubeService(ctx, deploymentSnapshot, kubeService)
	return err
}
