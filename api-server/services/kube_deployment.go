package services

import (
	"context"

	"k8s.io/utils/pointer"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type kubeDeploymentService struct{}

var KubeDeploymentService = kubeDeploymentService{}

func (s *kubeDeploymentService) MapKubeDeploymentToKubeDeploymentWithPods(ctx context.Context, deployment *appsv1.Deployment, pods []corev1.Pod, events []corev1.Event) *models.KubeDeploymentWithPods {
	filteredPods := make([]corev1.Pod, 0)

	for _, pod := range pods {
		match := true
		for k, v := range deployment.Spec.Selector.MatchLabels {
			if pod.Labels[k] != v {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		filteredPods = append(filteredPods, pod)
	}

	r := models.KubeDeploymentWithPods{
		Deployment: *deployment,
		Pods:       KubePodService.MapKubePodsToKubePodWithStatuses(ctx, filteredPods, events),
	}
	return &r
}

func (s *kubeDeploymentService) MapKubeDeploymentsToKubeDeploymentWithPodses(ctx context.Context, deployments []*appsv1.Deployment, pods []corev1.Pod, events []corev1.Event) []*models.KubeDeploymentWithPods {
	res := make([]*models.KubeDeploymentWithPods, 0, len(deployments))
	for _, d := range deployments {
		res = append(res, s.MapKubeDeploymentToKubeDeploymentWithPods(ctx, d, pods, events))
	}
	return res
}

func (s *kubeDeploymentService) DeploymentTargetToKubeDeployment(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeDeployment *appsv1.Deployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return
	}

	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	_, err = KubeNamespaceService.MakeSureNamespace(ctx, cluster, kubeNs)
	if err != nil {
		return
	}

	podTemplateSpec, err := KubePodService.DeploymentTargetToPodTemplateSpec(ctx, deploymentTarget)
	if err != nil {
		return
	}

	labels, err := DeploymentTargetService.GetKubeLabels(ctx, deploymentTarget)
	if err != nil {
		return
	}

	annotations, err := DeploymentTargetService.GetKubeAnnotations(ctx, deploymentTarget)
	if err != nil {
		return
	}

	kubeName, err := DeploymentTargetService.GetKubeName(ctx, deploymentTarget)
	if err != nil {
		return
	}

	defaultMaxSurge := intstr.FromString("25%")
	defaultMaxUnavailable := intstr.FromString("25%")

	strategy := appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       &defaultMaxSurge,
			MaxUnavailable: &defaultMaxUnavailable,
		},
	}

	kubeDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       kubeNs,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.KubeLabelYataiSelector: kubeName,
				},
			},
			Template: *podTemplateSpec,
			Strategy: strategy,
		},
		Status: appsv1.DeploymentStatus{},
	}
	return
}

func (s *kubeDeploymentService) DeployDeploymentTargetAsKubeDeployment(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) error {
	kubeDeployment, err := s.DeploymentTargetToKubeDeployment(ctx, deploymentTarget, deployOption)
	if err != nil {
		return errors.Wrap(err, "to k8s deployment")
	}
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return err
	}
	kubeDeploymentsCli, err := DeploymentService.GetKubeDeploymentsCli(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "get k8s deployments cli")
	}
	logrus.Infof("get k8s deployment %s ...", kubeDeployment.Name)
	_, err = kubeDeploymentsCli.Get(ctx, kubeDeployment.Name, metav1.GetOptions{})
	notFound := apierrors.IsNotFound(err)
	if !notFound && err != nil {
		return errors.Wrapf(err, "get k8s deployment %s", kubeDeployment.Name)
	}
	if notFound {
		logrus.Infof("create k8s deployment %s ...", kubeDeployment.Name)
		_, err = kubeDeploymentsCli.Create(ctx, kubeDeployment, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "create k8s deployment %s", kubeDeployment.Name)
		}
	} else {
		logrus.Infof("update k8s deployment %s ...", kubeDeployment.Name)
		_, err = kubeDeploymentsCli.Update(ctx, kubeDeployment, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "update k8s deployment %s", kubeDeployment.Name)
		}
	}

	return KubeHPAService.DeployDeploymentTargetAsKubeHPA(ctx, deploymentTarget, deployOption)
}
