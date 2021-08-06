package services

import (
	"context"
	"fmt"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/utils/pointer"
)

type kubePodService struct{}

var KubePodService = kubePodService{}

func (*kubePodService) ListPodsByDeployment(podLister v1.PodNamespaceLister, deployment *models.Deployment) ([]apiv1.Pod, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s = %d", consts.KubeLabelYataiDeploymentId, deployment.ID))
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
	return pods, nil
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
func (s *kubePodService) GetKubePodStatus(pod apiv1.Pod, warnings []apiv1.Event) models.KubePodStatus {
	var states []apiv1.ContainerState
	for _, containerStatus := range pod.Status.ContainerStatuses {
		states = append(states, containerStatus.State)
	}

	return models.KubePodStatus{
		Status:          string(s.getKubePodStatusPhase(pod, warnings)),
		Phase:           pod.Status.Phase,
		ContainerStates: states,
	}
}

// getKubePodStatusPhase returns one of four pod status phases (Pending, Running, Succeeded, Failed, Unknown, Terminating)
func (s *kubePodService) getKubePodStatusPhase(pod apiv1.Pod, warnings []apiv1.Event) apiv1.PodPhase {
	// For terminated pods that failed
	if pod.Status.Phase == apiv1.PodFailed {
		return apiv1.PodFailed
	}

	// For successfully terminated pods
	if pod.Status.Phase == apiv1.PodSucceeded {
		return apiv1.PodSucceeded
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
		return apiv1.PodRunning
	}

	// If the pod would otherwise be pending but has warning then label it as
	// failed and show and error to the user.
	if len(warnings) > 0 {
		return apiv1.PodFailed
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		return apiv1.PodUnknown
	} else if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	// pending
	return apiv1.PodPending
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

func getResourcesConfig(containerName string, resources *models.DeploymentResources, resourceMap map[string]*models.DeploymentResources, gpuNvidiaResourceRequest bool) (apiv1.ResourceRequirements, error) {
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
