package services

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/utils/pointer"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
)

type kubePodService struct{}

var KubePodService = kubePodService{}

func (s *kubePodService) ListPodsByDeployment(ctx context.Context, podLister v1.PodNamespaceLister, deployment *models.Deployment) ([]*models.KubePodWithStatus, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s = %s", commonconsts.KubeLabelYataiBentoDeployment, deployment.Name))
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
