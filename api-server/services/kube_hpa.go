package services

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type kubeHPAService struct{}

var KubeHPAService = kubeHPAService{}

func (s *kubeHPAService) DeploymentSnapshotToKubeHPA(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) (hpa *v2beta2.HorizontalPodAutoscaler, err error) {
	conf := deploymentSnapshot.Config
	if conf == nil {
		return
	}
	if conf.HPAConf == nil {
		return
	}

	labels, err := DeploymentSnapshotService.GetKubeLabels(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	annotations, err := DeploymentSnapshotService.GetKubeAnnotations(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	kubeName, err := DeploymentSnapshotService.GetKubeName(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	hpaConf := conf.HPAConf

	var metrics []v2beta2.MetricSpec
	if hpaConf.QPS != nil && *hpaConf.QPS > 0 {
		metrics = append(metrics, v2beta2.MetricSpec{
			Type: v2beta2.PodsMetricSourceType,
			Pods: &v2beta2.PodsMetricSource{
				Metric: v2beta2.MetricIdentifier{
					Name: consts.KubeHPAQPSMetric,
				},
				Target: v2beta2.MetricTarget{
					Type:         v2beta2.UtilizationMetricType,
					AverageValue: resource.NewQuantity(*hpaConf.QPS, resource.DecimalSI),
				},
			},
		})
	}

	if hpaConf.CPU != nil && *hpaConf.CPU > 0 {
		metrics = append(metrics, v2beta2.MetricSpec{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: v2beta2.MetricTarget{
					Type:               v2beta2.UtilizationMetricType,
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
		metrics = append(metrics, v2beta2.MetricSpec{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: v2beta2.MetricTarget{
					Type:         v2beta2.UtilizationMetricType,
					AverageValue: &quantity,
				},
			},
		})
	}

	maxReplicas := int32(consts.AppCompMaxReplicas)
	if hpaConf.MaxReplicas != nil {
		maxReplicas = *hpaConf.MaxReplicas
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	kubeHpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeName,
			Namespace:       kubeNs,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: deployOption.OwnerReferences,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MinReplicas: hpaConf.MinReplicas,
			MaxReplicas: maxReplicas,
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       kubeName,
			},
			Metrics: metrics,
		},
	}

	return kubeHpa, err
}

func (s *kubeHPAService) DeployDeploymentSnapshotAsKubeHPA(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, deployOption *models.DeployOption) error {
	kubeHpa, err := s.DeploymentSnapshotToKubeHPA(ctx, deploymentSnapshot, deployOption)
	if err != nil {
		return errors.Wrap(err, "failed convert comp to hpa failed")
	}
	kubeCli, _, err := DeploymentSnapshotService.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return errors.Wrap(err, "get kube cli set")
	}
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return errors.Wrap(err, "get deployment")
	}
	kubeNs := DeploymentService.GetKubeNamespace(deployment)
	hpaCli := kubeCli.AutoscalingV2beta2().HorizontalPodAutoscalers(kubeNs)
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

			patch, err := strategicpatch.CreateTwoWayMergePatch(oldJson, newJson, v2beta2.HorizontalPodAutoscaler{})
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
		kubeName, err := DeploymentSnapshotService.GetKubeName(ctx, deploymentSnapshot)
		if err != nil {
			return errors.Wrap(err, "get app comp kube name")
		}
		_ = hpaCli.Delete(ctx, kubeName, metav1.DeleteOptions{})
	}
	return nil
}
