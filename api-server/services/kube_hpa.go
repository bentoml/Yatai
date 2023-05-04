package services

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
)

type kubeHPAService struct{}

var KubeHPAService = kubeHPAService{}

func (s *kubeHPAService) DeploymentTargetToKubeHPA(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (hpa *v2.HorizontalPodAutoscaler, err error) {
	conf := deploymentTarget.Config
	if conf == nil {
		return
	}
	if conf.HPAConf == nil {
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

	hpaConf := conf.HPAConf

	var metrics []v2.MetricSpec
	if hpaConf.QPS != nil && *hpaConf.QPS > 0 {
		metrics = append(metrics, v2.MetricSpec{
			Type: v2.PodsMetricSourceType,
			Pods: &v2.PodsMetricSource{
				Metric: v2.MetricIdentifier{
					Name: commonconsts.KubeHPAQPSMetric,
				},
				Target: v2.MetricTarget{
					Type:         v2.UtilizationMetricType,
					AverageValue: resource.NewQuantity(*hpaConf.QPS, resource.DecimalSI),
				},
			},
		})
	}

	if hpaConf.CPU != nil && *hpaConf.CPU > 0 {
		metrics = append(metrics, v2.MetricSpec{
			Type: v2.ResourceMetricSourceType,
			Resource: &v2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: v2.MetricTarget{
					Type:               v2.UtilizationMetricType,
					AverageUtilization: hpaConf.CPU,
				},
			},
		})
	}

	if hpaConf.Memory != nil && *hpaConf.Memory != "" {
		quantity, err := resource.ParseQuantity(*hpaConf.Memory)
		if err != nil {
			return nil, errors.Wrapf(err, "parse memory %s", *hpaConf.Memory)
		}
		metrics = append(metrics, v2.MetricSpec{
			Type: v2.ResourceMetricSourceType,
			Resource: &v2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: v2.MetricTarget{
					Type:         v2.UtilizationMetricType,
					AverageValue: &quantity,
				},
			},
		})
	}

	if len(metrics) == 0 {
		metrics = append(metrics, v2.MetricSpec{
			Type: v2.ResourceMetricSourceType,
			Resource: &v2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: v2.MetricTarget{
					Type:               v2.UtilizationMetricType,
					AverageUtilization: utils.Int32Ptr(80),
				},
			},
		})
	}

	maxReplicas := int32(consts.AppCompMaxReplicas)
	if hpaConf.MaxReplicas != nil {
		maxReplicas = *hpaConf.MaxReplicas
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return nil, err
	}

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	kubeHpa := &v2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       kubeNs,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			MinReplicas: hpaConf.MinReplicas,
			MaxReplicas: maxReplicas,
			ScaleTargetRef: v2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       kubeName,
			},
			Metrics: metrics,
		},
	}

	return kubeHpa, err
}

func (s *kubeHPAService) DeployDeploymentTargetAsKubeHPA(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) error {
	kubeHpa, err := s.DeploymentTargetToKubeHPA(ctx, deploymentTarget, deployOption)
	if err != nil {
		return errors.Wrap(err, "failed convert comp to hpa failed")
	}
	kubeCli, _, err := DeploymentTargetService.GetKubeCliSet(ctx, deploymentTarget)
	if err != nil {
		return errors.Wrap(err, "get kube cli set")
	}
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}
	kubeNs := DeploymentService.GetKubeNamespace(deployment)
	hpaCli := kubeCli.AutoscalingV2().HorizontalPodAutoscalers(kubeNs)
	if err != nil {
		return errors.Wrap(err, "get KubeHPA cli failed")
	}
	if kubeHpa != nil {
		logrus.Infof("get k8s hpa %s ...", kubeHpa.Name)
		oldHPA, err := hpaCli.Get(ctx, kubeHpa.Name, metav1.GetOptions{})
		notFound := apierrors.IsNotFound(err)
		if !notFound && err != nil {
			return errors.Wrapf(err, "get k8s hpa %s", kubeHpa.Name)
		}
		if notFound {
			logrus.Infof("create k8s hpa %s ...", kubeHpa.Name)
			_, err = hpaCli.Create(ctx, kubeHpa, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "create k8s hpa %s", kubeHpa.Name)
			}
		} else {
			oldJson, err := json.Marshal(oldHPA)
			if err != nil {
				return errors.Wrap(err, "old hpa to json")
			}
			newJson, err := json.Marshal(kubeHpa)
			if err != nil {
				return errors.Wrap(err, "cur hpa to json")
			}

			patch, err := strategicpatch.CreateTwoWayMergePatch(oldJson, newJson, v2.HorizontalPodAutoscaler{})
			if err != nil {
				return errors.Wrap(err, "create json patch")
			}

			if len(patch) == 0 || string(patch) == "{}" {
				logrus.Infof("k8s hpa %s no modified ...", kubeHpa.Name)
				return nil
			}

			logrus.Infof("patch k8s hpa %s ...", kubeHpa.Name)
			_, err = hpaCli.Patch(ctx, kubeHpa.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "patch deployment")
			}
		}
	} else {
		kubeName, err := DeploymentTargetService.GetKubeName(ctx, deploymentTarget)
		if err != nil {
			return errors.Wrap(err, "get app comp kube name")
		}
		_ = hpaCli.Delete(ctx, kubeName, metav1.DeleteOptions{})
	}
	return nil
}
