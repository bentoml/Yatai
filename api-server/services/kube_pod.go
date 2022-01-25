package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bentoml/yatai/common/utils"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/utils/pointer"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type kubePodService struct{}

var KubePodService = kubePodService{}

func (s *kubePodService) ListPodsByDeployment(ctx context.Context, podLister v1.PodNamespaceLister, deployment *models.Deployment) ([]*models.KubePodWithStatus, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s = %s", consts.KubeLabelYataiDeployment, deployment.Name))
	if err != nil {
		return nil, err
	}
	pods_, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}
	pods := make([]apiv1.Pod, 0, len(pods_))
	for _, p := range pods_ {
		pods = append(pods, *p)
	}

	events, err := KubeEventService.ListAllKubeEventsByDeployment(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return s.MapKubePodsToKubePodWithStatuses(ctx, pods, events), nil
}

func (s *kubePodService) ListPodsBySelector(ctx context.Context, cluster *models.Cluster, namespace string, podLister v1.PodNamespaceLister, selector labels.Selector) ([]*models.KubePodWithStatus, error) {
	pods_, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}
	pods := make([]apiv1.Pod, 0, len(pods_))
	for _, p := range pods_ {
		pods = append(pods, *p)
	}

	events, err := KubeEventService.ListAllKubeEvents(ctx, cluster, namespace, func(event *apiv1.Event) bool {
		return true
	})
	if err != nil {
		return nil, err
	}

	return s.MapKubePodsToKubePodWithStatuses(ctx, pods, events), nil
}

func (s *kubePodService) MapKubePodsToKubePodWithStatuses(ctx context.Context, pods []apiv1.Pod, events []apiv1.Event) []*models.KubePodWithStatus {
	warningsMapping := KubeEventService.GetKubePodsWarningEventsMapping(events, pods)
	res := make([]*models.KubePodWithStatus, 0, len(pods))
	for _, pod := range pods {
		warnings := warningsMapping[pod.GetUID()]
		status := s.GetKubePodStatus(pod, warnings)
		res = append(res, &models.KubePodWithStatus{
			Pod:      pod,
			Status:   status,
			Warnings: warnings,
		})
	}
	return res
}

// GetKubePodRestartCount return the restart count of given pod (total number of its containers restarts).
func (s *kubePodService) GetKubePodRestartCount(pod apiv1.Pod) int32 {
	var restartCount int32 = 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	return restartCount
}

// GetKubePodStatus returns a KubePodStatus object containing a summary of the pod's status.
func (s *kubePodService) GetKubePodStatus(pod apiv1.Pod, warnings []apiv1.Event) modelschemas.KubePodStatus {
	var states []apiv1.ContainerState
	for _, containerStatus := range pod.Status.ContainerStatuses {
		states = append(states, containerStatus.State)
	}

	return modelschemas.KubePodStatus{
		Status:          s.getKubePodActualStatus(pod, warnings),
		Phase:           pod.Status.Phase,
		ContainerStates: states,
	}
}

// getKubePodActualStatus returns one of four pod status phases (Pending, Running, Succeeded, Failed, Unknown, Terminating)
func (s *kubePodService) getKubePodActualStatus(pod apiv1.Pod, warnings []apiv1.Event) modelschemas.KubePodActualStatus {
	// For terminated pods that failed
	if pod.Status.Phase == apiv1.PodFailed {
		return modelschemas.KubePodActualStatusFailed
	}

	// For successfully terminated pods
	if pod.Status.Phase == apiv1.PodSucceeded {
		return modelschemas.KubePodActualStatusSucceeded
	}

	ready := false
	initialized := false
	for _, c := range pod.Status.Conditions {
		if c.Type == apiv1.PodReady {
			ready = c.Status == apiv1.ConditionTrue
		}
		if c.Type == apiv1.PodInitialized {
			initialized = c.Status == apiv1.ConditionTrue
		}
	}

	if initialized && ready && pod.Status.Phase == apiv1.PodRunning {
		return modelschemas.KubePodActualStatusRunning
	}

	// If the pod would otherwise be pending but has warning then label it as
	// failed and show and error to the user.
	if len(warnings) > 0 {
		return modelschemas.KubePodActualStatusFailed
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		return modelschemas.KubePodActualStatusUnknown
	} else if pod.DeletionTimestamp != nil {
		return modelschemas.KubePodActualStatusTerminating
	}

	// pending
	return modelschemas.KubePodActualStatusPending
}

func (s *kubePodService) DeleteKubePod(ctx context.Context, deployment *models.Deployment, kubePodName string, force bool) error {
	podsCli, err := DeploymentService.GetKubePodsCli(ctx, deployment)
	if err != nil {
		return errors.Wrapf(err, "%s get k8s pods cli", deployment.Name)
	}
	var options metav1.DeleteOptions
	if force {
		policy := metav1.DeletePropagationForeground
		options = metav1.DeleteOptions{
			GracePeriodSeconds: pointer.Int64Ptr(0),
			PropagationPolicy:  &policy,
		}
	}
	logrus.Infof("delete k8s pod %s ...", kubePodName)
	return podsCli.Delete(ctx, kubePodName, options)
}

func (s *kubePodService) DeploymentTargetToPodTemplateSpec(ctx context.Context, deploymentTarget *models.DeploymentTarget) (podTemplateSpec *apiv1.PodTemplateSpec, err error) {
	podLabels, err := DeploymentTargetService.GetKubeLabels(ctx, deploymentTarget)
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

	bento, err := BentoService.GetAssociatedBento(ctx, deploymentTarget)
	if err != nil {
		return
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return
	}

	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}

	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return
	}

	dockerRegistry, err := OrganizationService.GetDockerRegistry(ctx, org)
	if err != nil {
		return
	}

	majorCluster, err := OrganizationService.GetMajorCluster(ctx, org)
	if err != nil {
		return
	}

	imageName, err := BentoService.GetImageName(ctx, bento, cluster.ID == majorCluster.ID)
	if err != nil {
		return
	}

	livenessProbe := &apiv1.Probe{
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
		FailureThreshold:    6,
		Handler: apiv1.Handler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path: "/livez",
				Port: intstr.FromInt(consts.BentoServicePort),
			},
		},
	}

	readinessProbe := &apiv1.Probe{
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
		FailureThreshold:    6,
		Handler: apiv1.Handler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path: "/readyz",
				Port: intstr.FromInt(consts.BentoServicePort),
			},
		},
	}

	containers := make([]apiv1.Container, 0, 1)

	vs := make([]apiv1.Volume, 0)
	vms := make([]apiv1.VolumeMount, 0)

	models_, _, err := ModelService.List(ctx, ListModelOption{
		BentoIds: &[]uint{bento.ID},
		Order:    utils.StringPtr("model.build_at ASC"),
	})
	if err != nil {
		return
	}

	args := make([]string, 0)
	imageTlsVerify := "false"
	if dockerRegistry.Secure {
		imageTlsVerify = "true"
	}

	for _, model := range models_ {
		var imageName_ string
		imageName_, err = ModelService.GetImageName(ctx, model, cluster.ID == majorCluster.ID)
		if err != nil {
			return
		}
		var modelRepository *models.ModelRepository
		modelRepository, err = ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
		if err != nil {
			return
		}
		pvName := fmt.Sprintf("pv-%s", model.Version)
		sourcePath := fmt.Sprintf("/models/%s/%s", modelRepository.Name, model.Version)
		destDirPath := fmt.Sprintf("./models/%s", modelRepository.Name)
		destPath := filepath.Join(destDirPath, model.Version)
		args = append(args, "mkdir", "-p", destDirPath, ";", "ln", "-sf", filepath.Join(sourcePath, "model"), destPath, ";", "echo", "-n", fmt.Sprintf("'%s'", model.Version), ">", filepath.Join(destDirPath, "latest"), ";")
		v := apiv1.Volume{
			Name: pvName,
			VolumeSource: apiv1.VolumeSource{
				CSI: &apiv1.CSIVolumeSource{
					Driver: consts.KubeCSIDriverImage,
					VolumeAttributes: map[string]string{
						"image":     imageName_,
						"tlsVerify": imageTlsVerify,
					},
				},
			},
		}
		vs = append(vs, v)
		vm := apiv1.VolumeMount{
			Name:      pvName,
			MountPath: sourcePath,
		}
		vms = append(vms, vm)
	}

	var envs []apiv1.EnvVar
	if deploymentTarget.Config != nil && deploymentTarget.Config.Envs != nil {
		envs = make([]apiv1.EnvVar, 0, len(*deploymentTarget.Config.Envs))
		for _, v := range *deploymentTarget.Config.Envs {
			envs = append(envs, apiv1.EnvVar{
				Name:  v.Key,
				Value: v.Value,
			})
		}
	}

	args = append(args, "./env/docker/entrypoint.sh", "bentoml", "serve", ".", "--production")

	container := apiv1.Container{
		Name:           kubeName,
		Image:          imageName,
		Command:        []string{"sh", "-c"},
		Args:           []string{strings.Join(args, " ")},
		LivenessProbe:  livenessProbe,
		ReadinessProbe: readinessProbe,
		Env:            envs,
		TTY:            true,
		Stdin:          true,
		VolumeMounts:   vms,
	}

	containers = append(containers, container)

	podLabels[consts.KubeLabelYataiSelector] = kubeName

	podTemplateSpec = &apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      podLabels,
			Annotations: annotations,
		},
		Spec: apiv1.PodSpec{
			Containers: containers,
			Volumes:    vs,
		},
	}

	if dockerRegistry.Username != "" {
		podTemplateSpec.Spec.ImagePullSecrets = []apiv1.LocalObjectReference{
			{
				Name: consts.KubeSecretNameRegcred,
			},
		}
	}

	return
}

// nolint:unused,deadcode
func getResourcesConfig(containerName string, resources *modelschemas.DeploymentTargetResources, resourceMap map[string]*modelschemas.DeploymentTargetResources, gpuNvidiaResourceRequest bool) (apiv1.ResourceRequirements, error) {
	currentResources := apiv1.ResourceRequirements{
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse("300m"),
			apiv1.ResourceMemory: resource.MustParse("500Mi"),
		},
		Limits: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse("500m"),
			apiv1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}
	if gpuNvidiaResourceRequest {
		currentResources.Limits[consts.KubeResourceGPUNvidia] = resource.MustParse("1")
	}

	resourceConf := resources
	if resourceMap != nil {
		if _, ok := resourceMap[containerName]; ok {
			resourceConf = resourceMap[containerName]
		}
	}
	if resourceConf != nil {
		if resourceConf.Limits != nil {
			if resourceConf.Limits.CPU != "" {
				q, err := resource.ParseQuantity(resourceConf.Limits.CPU)
				if err != nil {
					return currentResources, errors.Wrapf(err, "parse limits cpu quantity")
				}
				if currentResources.Limits == nil {
					currentResources.Limits = make(apiv1.ResourceList)
				}
				currentResources.Limits[apiv1.ResourceCPU] = q
			}
			if resourceConf.Limits.Memory != "" {
				q, err := resource.ParseQuantity(resourceConf.Limits.Memory)
				if err != nil {
					return currentResources, errors.Wrapf(err, "parse limits memory quantity")
				}
				if currentResources.Limits == nil {
					currentResources.Limits = make(apiv1.ResourceList)
				}
				currentResources.Limits[apiv1.ResourceMemory] = q
			}
			if resourceConf.Limits.GPU != "" {
				q, err := resource.ParseQuantity(resourceConf.Limits.GPU)
				if err != nil {
					return currentResources, errors.Wrapf(err, "parse limits gpu quantity")
				}
				if currentResources.Limits == nil {
					currentResources.Limits = make(apiv1.ResourceList)
				}
				currentResources.Limits[consts.KubeResourceGPUNvidia] = q
			}
		}
		if resourceConf.Requests != nil {
			if resourceConf.Requests.CPU != "" {
				q, err := resource.ParseQuantity(resourceConf.Requests.CPU)
				if err != nil {
					return currentResources, errors.Wrapf(err, "parse requests cpu quantity")
				}
				if currentResources.Requests == nil {
					currentResources.Requests = make(apiv1.ResourceList)
				}
				currentResources.Requests[apiv1.ResourceCPU] = q
			}
			if resourceConf.Requests.Memory != "" {
				q, err := resource.ParseQuantity(resourceConf.Requests.Memory)
				if err != nil {
					return currentResources, errors.Wrapf(err, "parse requests memory quantity")
				}
				if currentResources.Requests == nil {
					currentResources.Requests = make(apiv1.ResourceList)
				}
				currentResources.Requests[apiv1.ResourceMemory] = q
			}
		}
	}
	return currentResources, nil
}
