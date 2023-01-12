package services

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"

	servingconversion "github.com/bentoml/yatai-deployment/apis/serving/conversion"
	servingv1alpha2 "github.com/bentoml/yatai-deployment/apis/serving/v1alpha2"
	servingv1alpha3 "github.com/bentoml/yatai-deployment/apis/serving/v1alpha3"
	servingv2alpha1 "github.com/bentoml/yatai-deployment/apis/serving/v2alpha1"

	resourcesv1alpha1 "github.com/bentoml/yatai-image-builder/apis/resources/v1alpha1"
)

type kubeBentoDeploymentService struct{}

var KubeBentoDeploymentService = kubeBentoDeploymentService{}

func (s *kubeBentoDeploymentService) transformToBentoDeploymentV1alpha2(ctx context.Context, deploymentTarget *models.DeploymentTarget) (kubeBentoDeployment *servingv1alpha2.BentoDeployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated deployment")
		return
	}

	bento, err := BentoService.GetAssociatedBento(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated bento")
		return
	}
	tag, err := BentoService.GetTag(ctx, bento)
	if err != nil {
		err = errors.Wrap(err, "failed to get bento tag")
		return
	}

	var autoscalingSpec *modelschemas.DeploymentTargetHPAConf
	if deploymentTarget.Config != nil {
		autoscalingSpec = deploymentTarget.Config.HPAConf
	}

	envs := make([]modelschemas.LabelItemSchema, 0)
	if deploymentTarget.Config != nil && deploymentTarget.Config.Envs != nil {
		for _, env := range *deploymentTarget.Config.Envs {
			envs = append(envs, *env)
		}
	}

	var resources *modelschemas.DeploymentTargetResources
	if deploymentTarget.Config != nil {
		resources = deploymentTarget.Config.Resources
	}

	var runners []servingv1alpha2.BentoDeploymentRunnerSpec
	if deploymentTarget.Config != nil && deploymentTarget.Config.Runners != nil {
		runners = make([]servingv1alpha2.BentoDeploymentRunnerSpec, 0, len(deploymentTarget.Config.Runners))
		for name, runner := range deploymentTarget.Config.Runners {
			envs_ := make([]modelschemas.LabelItemSchema, 0)
			if runner.Envs != nil {
				for _, env := range *runner.Envs {
					envs_ = append(envs_, *env)
				}
			}
			runners = append(runners, servingv1alpha2.BentoDeploymentRunnerSpec{
				Name:        name,
				Resources:   servingconversion.ConvertFromDeploymentTargetResources(runner.Resources),
				Autoscaling: runner.HPAConf,
				Envs:        &envs_,
			})
		}
	}

	ingress := servingv1alpha2.BentoDeploymentIngressSpec{}

	if deploymentTarget.Config != nil && deploymentTarget.Config.EnableIngress != nil && *deploymentTarget.Config.EnableIngress {
		ingress.Enabled = true
	}

	kubeBentoDeployment = &servingv1alpha2.BentoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: DeploymentService.GetKubeNamespace(deployment),
		},
		Spec: servingv1alpha2.BentoDeploymentSpec{
			BentoTag:    string(tag),
			Autoscaling: autoscalingSpec,
			Envs:        &envs,
			Resources:   servingconversion.ConvertFromDeploymentTargetResources(resources),
			Runners:     runners,
			Ingress:     ingress,
		},
	}

	return
}

func (s *kubeBentoDeploymentService) transformToBentoDeploymentV1alpha3(ctx context.Context, deploymentTarget *models.DeploymentTarget) (kubeBentoDeployment *servingv1alpha3.BentoDeployment, err error) {
	kubeBentoDeploymentV1alpha2, err := s.transformToBentoDeploymentV1alpha2(ctx, deploymentTarget)
	if err != nil {
		return
	}
	kubeBentoDeployment = kubeBentoDeploymentV1alpha2.ConvertToV1alpha3()
	if kubeBentoDeployment == nil {
		return
	}

	if kubeBentoDeployment.Spec.Annotations == nil {
		kubeBentoDeployment.Spec.Annotations = make(map[string]string)
	}

	if deploymentTarget.Config != nil {
		if deploymentTarget.Config.EnableDebugMode != nil && *deploymentTarget.Config.EnableDebugMode {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueTrue
		} else {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueFalse
		}
		if deploymentTarget.Config.EnableStealingTrafficDebugMode != nil && *deploymentTarget.Config.EnableStealingTrafficDebugMode {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueTrue
		} else {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueFalse
		}
		if deploymentTarget.Config.EnableDebugPodReceiveProductionTraffic != nil && *deploymentTarget.Config.EnableDebugPodReceiveProductionTraffic {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueTrue
		} else {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueFalse
		}
		if deploymentTarget.Config.DeploymentStrategy != nil {
			kubeBentoDeployment.Spec.Annotations[KubeAnnotationDeploymentStrategy] = string(*deploymentTarget.Config.DeploymentStrategy)
		}
		if deploymentTarget.Config.BentoDeploymentOverrides != nil {
			kubeBentoDeployment.Spec.ExtraPodMetadata = servingv1alpha3.TransformToOldExtraPodMetadata(deploymentTarget.Config.BentoDeploymentOverrides.ExtraPodMetadata)
			kubeBentoDeployment.Spec.ExtraPodSpec = servingv1alpha3.TransformToOldExtraPodSpec(deploymentTarget.Config.BentoDeploymentOverrides.ExtraPodSpec)
		}
		for name, runner := range deploymentTarget.Config.Runners {
			if runner.BentoDeploymentOverrides != nil {
				for _, runner_ := range kubeBentoDeployment.Spec.Runners {
					if runner_.Name == name {
						runner_.ExtraPodMetadata = servingv1alpha3.TransformToOldExtraPodMetadata(runner.BentoDeploymentOverrides.ExtraPodMetadata)
						runner_.ExtraPodSpec = servingv1alpha3.TransformToOldExtraPodSpec(runner.BentoDeploymentOverrides.ExtraPodSpec)
					}
				}
			}
		}
	}
	return
}

func (s *kubeBentoDeploymentService) transformToBentoDeploymentV2alpha1(ctx context.Context, deploymentTarget *models.DeploymentTarget) (kubeBentoDeployment *servingv2alpha1.BentoDeployment, bentoRequest *resourcesv1alpha1.BentoRequest, err error) {
	kubeBentoDeployment_, err := s.transformToBentoDeploymentV1alpha3(ctx, deploymentTarget)
	if err != nil {
		return
	}
	bentoRequest = kubeBentoDeployment_.ConvertToBentoRequest()
	kubeBentoDeployment = &servingv2alpha1.BentoDeployment{}
	err = kubeBentoDeployment_.ConvertToV2alpha1(kubeBentoDeployment, bentoRequest.Name)
	if err != nil {
		return
	}
	if deploymentTarget.Config != nil {
		if deploymentTarget.Config.BentoRequestOverrides != nil {
			bentoRequest.Spec.ImageBuildTimeout = deploymentTarget.Config.BentoRequestOverrides.ImageBuildTimeout
			bentoRequest.Spec.ImageBuilderExtraPodMetadata = deploymentTarget.Config.BentoRequestOverrides.ImageBuilderExtraPodMetadata
			bentoRequest.Spec.ImageBuilderExtraPodSpec = deploymentTarget.Config.BentoRequestOverrides.ImageBuilderExtraPodSpec
			bentoRequest.Spec.ImageBuilderExtraContainerEnv = deploymentTarget.Config.BentoRequestOverrides.ImageBuilderExtraContainerEnv
			bentoRequest.Spec.ImageBuilderContainerResources = deploymentTarget.Config.BentoRequestOverrides.ImageBuilderContainerResources
			bentoRequest.Spec.DockerConfigJSONSecretName = deploymentTarget.Config.BentoRequestOverrides.DockerConfigJSONSecretName
			bentoRequest.Spec.DownloaderContainerEnvFrom = deploymentTarget.Config.BentoRequestOverrides.DownloaderContainerEnvFrom
		}
		if deploymentTarget.Config.BentoDeploymentOverrides != nil {
			kubeBentoDeployment.Spec.MonitorExporter = deploymentTarget.Config.BentoDeploymentOverrides.MonitorExporter
			kubeBentoDeployment.Spec.ExtraPodMetadata = deploymentTarget.Config.BentoDeploymentOverrides.ExtraPodMetadata
			kubeBentoDeployment.Spec.ExtraPodSpec = deploymentTarget.Config.BentoDeploymentOverrides.ExtraPodSpec
		}
		for name, runner := range deploymentTarget.Config.Runners {
			if runner.BentoDeploymentOverrides != nil {
				for _, runner_ := range kubeBentoDeployment.Spec.Runners {
					if runner_.Name == name {
						runner_.ExtraPodMetadata = runner.BentoDeploymentOverrides.ExtraPodMetadata
						runner_.ExtraPodSpec = runner.BentoDeploymentOverrides.ExtraPodSpec
					}
				}
			}
		}
	}
	return
}

func (s *kubeBentoDeploymentService) DeployV1alpha2(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeBentoDeployment *servingv1alpha2.BentoDeployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated deployment")
		return
	}

	cli, err := DeploymentService.GetKubeBentoDeploymentV1alpha2Cli(ctx, deployment)
	if err != nil {
		err = errors.Wrap(err, "failed to get kube bento deployment cli")
		return
	}

	if deploymentTarget.Config != nil && deploymentTarget.Config.KubeResourceVersion != "" {
		var oldKubeBentoDeployment *servingv1alpha2.BentoDeployment
		oldKubeBentoDeployment, err = cli.Get(ctx, deployment.Name, metav1.GetOptions{})
		isNotFound := apierrors.IsNotFound(err)
		if err != nil && !isNotFound {
			err = errors.Wrap(err, "failed to get kube bento deployment")
			return
		}
		if !isNotFound && oldKubeBentoDeployment.ResourceVersion == deploymentTarget.Config.KubeResourceVersion {
			kubeBentoDeployment = oldKubeBentoDeployment
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusImageBuilding
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		deployment.Status = status
		ctx_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()
			_, _ = DeploymentService.SyncStatus(ctx_, deployment)
		}()
	}()

	kubeBentoDeployment, err = s.transformToBentoDeploymentV1alpha2(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to transform to kube bento deployment")
		return
	}

	var oldKubeBentoDeployment *servingv1alpha2.BentoDeployment
	oldKubeBentoDeployment, err = cli.Get(ctx, kubeBentoDeployment.Name, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		err = errors.Wrap(err, "failed to get kube bento deployment")
		return
	}
	if isNotFound {
		kubeBentoDeployment, err = cli.Create(ctx, kubeBentoDeployment, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to create kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	} else {
		kubeBentoDeployment.SetResourceVersion(oldKubeBentoDeployment.GetResourceVersion())
		if kubeBentoDeployment.Annotations == nil {
			kubeBentoDeployment.Annotations = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Annotations {
			if _, ok := kubeBentoDeployment.Annotations[k]; !ok {
				kubeBentoDeployment.Annotations[k] = v
			}
		}
		if kubeBentoDeployment.Labels == nil {
			kubeBentoDeployment.Labels = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Labels {
			if _, ok := kubeBentoDeployment.Labels[k]; !ok {
				kubeBentoDeployment.Labels[k] = v
			}
		}
		kubeBentoDeployment.Spec.Autoscaling = oldKubeBentoDeployment.Spec.Autoscaling
		for idx, runner := range kubeBentoDeployment.Spec.Runners {
			for _, oldRunner := range oldKubeBentoDeployment.Spec.Runners {
				if runner.Name == oldRunner.Name {
					kubeBentoDeployment.Spec.Runners[idx].Autoscaling = oldRunner.Autoscaling
				}
			}
		}
		kubeBentoDeployment, err = cli.Update(ctx, kubeBentoDeployment, metav1.UpdateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to update kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	}
	return
}

const (
	KubeAnnotationEnableDebugMode                        = "yatai.ai/enable-debug-mode"
	KubeAnnotationEnableStealingTrafficDebugMode         = "yatai.ai/enable-stealing-traffic-debug-mode"
	KubeAnnotationEnableDebugPodReceiveProductionTraffic = "yatai.ai/enable-debug-pod-receive-production-traffic"
	KubeAnnotationDeploymentStrategy                     = "yatai.ai/deployment-strategy"
)

func (s *kubeBentoDeploymentService) DeployV1alpha3(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeBentoDeployment *servingv1alpha3.BentoDeployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated deployment")
		return
	}

	cli, err := DeploymentService.GetKubeBentoDeploymentV1alpha3Cli(ctx, deployment)
	if err != nil {
		err = errors.Wrap(err, "failed to get kube bento deployment cli")
		return
	}

	if deploymentTarget.Config != nil && deploymentTarget.Config.KubeResourceVersion != "" {
		var oldKubeBentoDeployment *servingv1alpha3.BentoDeployment
		oldKubeBentoDeployment, err = cli.Get(ctx, deployment.Name, metav1.GetOptions{})
		isNotFound := apierrors.IsNotFound(err)
		if err != nil && !isNotFound {
			err = errors.Wrap(err, "failed to get kube bento deployment")
			return
		}
		if !isNotFound && oldKubeBentoDeployment.ResourceVersion == deploymentTarget.Config.KubeResourceVersion {
			kubeBentoDeployment = oldKubeBentoDeployment
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusImageBuilding
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		deployment.Status = status
		ctx_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()
			_, _ = DeploymentService.SyncStatus(ctx_, deployment)
		}()
	}()

	kubeBentoDeployment, err = s.transformToBentoDeploymentV1alpha3(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to transform to kube bento deployment")
		return
	}

	var oldKubeBentoDeployment *servingv1alpha3.BentoDeployment
	oldKubeBentoDeployment, err = cli.Get(ctx, kubeBentoDeployment.Name, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		err = errors.Wrap(err, "failed to get kube bento deployment")
		return
	}
	if isNotFound {
		kubeBentoDeployment, err = cli.Create(ctx, kubeBentoDeployment, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to create kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	} else {
		kubeBentoDeployment.SetResourceVersion(oldKubeBentoDeployment.GetResourceVersion())
		if kubeBentoDeployment.Annotations == nil {
			kubeBentoDeployment.Annotations = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Annotations {
			if _, ok := kubeBentoDeployment.Annotations[k]; !ok {
				kubeBentoDeployment.Annotations[k] = v
			}
		}
		if kubeBentoDeployment.Labels == nil {
			kubeBentoDeployment.Labels = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Labels {
			if _, ok := kubeBentoDeployment.Labels[k]; !ok {
				kubeBentoDeployment.Labels[k] = v
			}
		}
		// copy old spec annotations
		if kubeBentoDeployment.Spec.Annotations == nil {
			kubeBentoDeployment.Spec.Annotations = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Spec.Annotations {
			if _, ok := kubeBentoDeployment.Spec.Annotations[k]; !ok {
				kubeBentoDeployment.Spec.Annotations[k] = v
			}
		}
		kubeBentoDeployment.Spec.Labels = oldKubeBentoDeployment.Spec.Labels
		kubeBentoDeployment.Spec.Ingress.Annotations = oldKubeBentoDeployment.Spec.Ingress.Annotations
		kubeBentoDeployment.Spec.Ingress.Labels = oldKubeBentoDeployment.Spec.Ingress.Labels
		kubeBentoDeployment.Spec.Ingress.TLS = oldKubeBentoDeployment.Spec.Ingress.TLS
		for idx, runner := range kubeBentoDeployment.Spec.Runners {
			var runnerConfig *modelschemas.DeploymentTargetRunnerConfig
			if deploymentTarget.Config != nil {
				runnerConfig_ := deploymentTarget.Config.Runners[runner.Name]
				runnerConfig = &runnerConfig_
			}
			for _, oldRunner := range oldKubeBentoDeployment.Spec.Runners {
				if runner.Name == oldRunner.Name {
					if kubeBentoDeployment.Spec.Runners[idx].Annotations == nil {
						kubeBentoDeployment.Spec.Runners[idx].Annotations = make(map[string]string)
					}
					if runnerConfig != nil {
						if runnerConfig.EnableDebugMode != nil && *runnerConfig.EnableDebugMode {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueFalse
						}
						if runnerConfig.EnableStealingTrafficDebugMode != nil && *runnerConfig.EnableStealingTrafficDebugMode {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueFalse
						}
						if runnerConfig.EnableDebugPodReceiveProductionTraffic != nil && *runnerConfig.EnableDebugPodReceiveProductionTraffic {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueFalse
						}
						if runnerConfig.DeploymentStrategy != nil {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationDeploymentStrategy] = string(*runnerConfig.DeploymentStrategy)
						}
					}
					for k, v := range oldRunner.Annotations {
						if _, ok := kubeBentoDeployment.Spec.Runners[idx].Annotations[k]; !ok {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[k] = v
						}
					}
					kubeBentoDeployment.Spec.Runners[idx].Labels = oldRunner.Labels
				}
			}
		}
		kubeBentoDeployment, err = cli.Update(ctx, kubeBentoDeployment, metav1.UpdateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to update kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	}
	return
}

func (s *kubeBentoDeploymentService) DeployV2alpha1(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeBentoDeployment *servingv2alpha1.BentoDeployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated deployment")
		return
	}

	bentoRequestCli, err := DeploymentService.GetKubeBentoRequestV1alpha1Cli(ctx, deployment)
	if err != nil {
		err = errors.Wrap(err, "failed to get kube bento request cli")
		return
	}

	cli, err := DeploymentService.GetKubeBentoDeploymentV2alpha1Cli(ctx, deployment)
	if err != nil {
		err = errors.Wrap(err, "failed to get kube bento deployment cli")
		return
	}

	if deploymentTarget.Config != nil && deploymentTarget.Config.KubeResourceVersion != "" {
		var oldKubeBentoDeployment *servingv2alpha1.BentoDeployment
		oldKubeBentoDeployment, err = cli.Get(ctx, deployment.Name, metav1.GetOptions{})
		isNotFound := apierrors.IsNotFound(err)
		if err != nil && !isNotFound {
			err = errors.Wrap(err, "failed to get kube bento deployment")
			return
		}
		if !isNotFound && oldKubeBentoDeployment.ResourceVersion == deploymentTarget.Config.KubeResourceVersion {
			kubeBentoDeployment = oldKubeBentoDeployment
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusImageBuilding
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		deployment.Status = status
		ctx_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()
			_, _ = DeploymentService.SyncStatus(ctx_, deployment)
		}()
	}()

	kubeBentoDeployment, bentoRequest, err := s.transformToBentoDeploymentV2alpha1(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to transform to kube bento deployment")
		return
	}

	var oldBentoRequest *resourcesv1alpha1.BentoRequest
	oldBentoRequest, err = bentoRequestCli.Get(ctx, bentoRequest.Name, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		err = errors.Wrap(err, "failed to get kube bento request")
		return
	}
	if isNotFound {
		_, err = bentoRequestCli.Create(ctx, bentoRequest, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrap(err, "failed to create kube bento request")
			return
		}
	} else {
		bentoRequest.SetResourceVersion(oldBentoRequest.GetResourceVersion())
		_, err = bentoRequestCli.Update(ctx, bentoRequest, metav1.UpdateOptions{})
		if err != nil {
			err = errors.Wrap(err, "failed to update kube bento request")
			return
		}
	}

	var oldKubeBentoDeployment *servingv2alpha1.BentoDeployment
	oldKubeBentoDeployment, err = cli.Get(ctx, kubeBentoDeployment.Name, metav1.GetOptions{})
	isNotFound = apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		err = errors.Wrap(err, "failed to get kube bento deployment")
		return
	}
	if isNotFound {
		kubeBentoDeployment, err = cli.Create(ctx, kubeBentoDeployment, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to create kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	} else {
		kubeBentoDeployment.SetResourceVersion(oldKubeBentoDeployment.GetResourceVersion())
		if kubeBentoDeployment.Annotations == nil {
			kubeBentoDeployment.Annotations = map[string]string{}
		}
		for k, v := range oldKubeBentoDeployment.Annotations {
			if _, ok := kubeBentoDeployment.Annotations[k]; !ok {
				kubeBentoDeployment.Annotations[k] = v
			}
		}
		if kubeBentoDeployment.Labels == nil {
			kubeBentoDeployment.Labels = map[string]string{}
		}
		for k, v := range oldKubeBentoDeployment.Labels {
			if _, ok := kubeBentoDeployment.Labels[k]; !ok {
				kubeBentoDeployment.Labels[k] = v
			}
		}
		// copy old spec annotations
		if kubeBentoDeployment.Spec.Annotations == nil {
			kubeBentoDeployment.Spec.Annotations = make(map[string]string)
		}
		for k, v := range oldKubeBentoDeployment.Spec.Annotations {
			if _, ok := kubeBentoDeployment.Spec.Annotations[k]; !ok {
				kubeBentoDeployment.Spec.Annotations[k] = v
			}
		}
		kubeBentoDeployment.Spec.Labels = oldKubeBentoDeployment.Spec.Labels
		kubeBentoDeployment.Spec.Ingress.Annotations = oldKubeBentoDeployment.Spec.Ingress.Annotations
		kubeBentoDeployment.Spec.Ingress.Labels = oldKubeBentoDeployment.Spec.Ingress.Labels
		kubeBentoDeployment.Spec.Ingress.TLS = oldKubeBentoDeployment.Spec.Ingress.TLS
		currentAutoscaling := kubeBentoDeployment.Spec.Autoscaling
		kubeBentoDeployment.Spec.Autoscaling = oldKubeBentoDeployment.Spec.Autoscaling
		if currentAutoscaling != nil {
			if kubeBentoDeployment.Spec.Autoscaling == nil {
				kubeBentoDeployment.Spec.Autoscaling = currentAutoscaling
			} else {
				kubeBentoDeployment.Spec.Autoscaling.MinReplicas = currentAutoscaling.MinReplicas
				kubeBentoDeployment.Spec.Autoscaling.MaxReplicas = currentAutoscaling.MaxReplicas
			}
		}
		for idx, runner := range kubeBentoDeployment.Spec.Runners {
			var runnerConfig *modelschemas.DeploymentTargetRunnerConfig
			if deploymentTarget.Config != nil {
				runnerConfig_ := deploymentTarget.Config.Runners[runner.Name]
				runnerConfig = &runnerConfig_
			}
			for _, oldRunner := range oldKubeBentoDeployment.Spec.Runners {
				if runner.Name == oldRunner.Name {
					if kubeBentoDeployment.Spec.Runners[idx].Annotations == nil {
						kubeBentoDeployment.Spec.Runners[idx].Annotations = make(map[string]string)
					}
					if runnerConfig != nil {
						if runnerConfig.EnableDebugMode != nil && *runnerConfig.EnableDebugMode {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugMode] = commonconsts.KubeLabelValueFalse
						}
						if runnerConfig.EnableStealingTrafficDebugMode != nil && *runnerConfig.EnableStealingTrafficDebugMode {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableStealingTrafficDebugMode] = commonconsts.KubeLabelValueFalse
						}
						if runnerConfig.EnableDebugPodReceiveProductionTraffic != nil && *runnerConfig.EnableDebugPodReceiveProductionTraffic {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueTrue
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[KubeAnnotationEnableDebugPodReceiveProductionTraffic] = commonconsts.KubeLabelValueFalse
						}
					}
					for k, v := range oldRunner.Annotations {
						if _, ok := kubeBentoDeployment.Spec.Runners[idx].Annotations[k]; !ok {
							kubeBentoDeployment.Spec.Runners[idx].Annotations[k] = v
						}
					}
					kubeBentoDeployment.Spec.Runners[idx].Labels = oldRunner.Labels
					currentAutoscaling := kubeBentoDeployment.Spec.Runners[idx].Autoscaling
					kubeBentoDeployment.Spec.Runners[idx].Autoscaling = oldRunner.Autoscaling
					if currentAutoscaling != nil {
						if kubeBentoDeployment.Spec.Runners[idx].Autoscaling == nil {
							kubeBentoDeployment.Spec.Runners[idx].Autoscaling = currentAutoscaling
						} else {
							kubeBentoDeployment.Spec.Runners[idx].Autoscaling.MinReplicas = currentAutoscaling.MinReplicas
							kubeBentoDeployment.Spec.Runners[idx].Autoscaling.MaxReplicas = currentAutoscaling.MaxReplicas
						}
					}
				}
			}
		}
		kubeBentoDeployment, err = cli.Update(ctx, kubeBentoDeployment, metav1.UpdateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to update kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	}
	return
}
