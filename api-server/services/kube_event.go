package services

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

var KubeEventFailedReasonPartials = []string{"failed", "err", "exceeded", "invalid", "unhealthy",
	"mismatch", "insufficient", "conflict", "outof", "nil", "backoff"}

type kubeEventService struct{}

var KubeEventService = kubeEventService{}

func (s *kubeEventService) isKubePodReadyOrSucceeded(pod apiv1.Pod) bool {
	if pod.Status.Phase == apiv1.PodSucceeded {
		return true
	}
	if pod.Status.Phase == apiv1.PodRunning {
		for _, c := range pod.Status.Conditions {
			if c.Type == apiv1.PodReady {
				if c.Status == apiv1.ConditionFalse {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (s *kubeEventService) removeDuplicateKubeEvents(slice []apiv1.Event) []apiv1.Event {
	visited := make(map[string]bool)
	result := make([]apiv1.Event, 0)

	for _, elem := range slice {
		if !visited[elem.Reason] {
			visited[elem.Reason] = true
			result = append(result, elem)
		}
	}

	return result
}

func (s *kubeEventService) filterKubeEventsByPodsUID(events []apiv1.Event, pods []apiv1.Pod) []apiv1.Event {
	result := make([]apiv1.Event, 0)
	podEventMap := make(map[types.UID]bool)
	deploymentEventMap := make(map[string]bool)
	if len(pods) == 0 || len(events) == 0 {
		return result
	}

	for _, pod := range pods {
		podEventMap[pod.UID] = true
		if selector, exist := pod.Labels[consts.KubeLabelYataiSelector]; exist {
			deploymentEventMap[selector] = true
		}
	}

	for _, event := range events {
		if _, exists := podEventMap[event.InvolvedObject.UID]; exists && event.InvolvedObject.Kind == consts.KubeEventResourceKindPod {
			result = append(result, event)
			continue
		}

		if _, exists := deploymentEventMap[event.InvolvedObject.Name]; exists && event.InvolvedObject.Kind == consts.KubeEventResourceKindHPA {
			result = append(result, event)
			continue
		}

		if event.InvolvedObject.Kind == consts.KubeEventResourceKindReplicaSet {
			endIndex := strings.LastIndex(event.InvolvedObject.Name, "-")
			if endIndex == -1 {
				continue
			}
			deploymentName := event.InvolvedObject.Name[:endIndex]
			if _, exists := deploymentEventMap[deploymentName]; exists {
				result = append(result, event)
				continue
			}
		}
	}

	return result
}

func (s *kubeEventService) filterKubeEventsByType(events []apiv1.Event, eventType string) []apiv1.Event {
	if len(eventType) == 0 || len(events) == 0 {
		return events
	}

	result := make([]apiv1.Event, 0)
	for _, event := range events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}

	return result
}

func (s *kubeEventService) FilterWarningKubeEvents(events []apiv1.Event) []apiv1.Event {
	return s.filterKubeEventsByType(s.FillKubeEventsType(events), apiv1.EventTypeWarning)
}

func (s *kubeEventService) FillKubeEventsType(events []apiv1.Event) []apiv1.Event {
	for i := range events {
		// Fill in only events with empty type.
		if len(events[i].Type) == 0 {
			if s.isKubeEventFailedReason(events[i].Reason, KubeEventFailedReasonPartials...) {
				events[i].Type = apiv1.EventTypeWarning
			} else {
				events[i].Type = apiv1.EventTypeNormal
			}
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[j].LastTimestamp.After(events[i].LastTimestamp.Time)
	})

	return events
}

func (s *kubeEventService) isKubeEventFailedReason(reason string, partials ...string) bool {
	for _, partial := range partials {
		if strings.Contains(strings.ToLower(reason), partial) {
			return true
		}
	}

	return false
}

func (s *kubeEventService) ListAllKubeEventsByDeployment(ctx context.Context, deployment *models.Deployment) ([]apiv1.Event, error) {
	return s.ListAllKubeEventsByDeploymentSnapshot(ctx, deployment, nil)
}

func (s *kubeEventService) MakeKubeEventFilter(ctx context.Context, deployment *models.Deployment, deploymentSnapshot **models.DeploymentSnapshot) (func(event *apiv1.Event) bool, error) {
	var err error
	var kubeName string

	if deploymentSnapshot != nil {
		kubeName, err = DeploymentSnapshotService.GetKubeName(ctx, *deploymentSnapshot)
		if err != nil {
			return nil, err
		}
	} else {
		kubeName = DeploymentService.GetKubeName(deployment)
	}

	kubeNamePattern, err := regexp.Compile(fmt.Sprintf("^%s-", kubeName))
	if err != nil {
		return nil, errors.Wrap(err, "compile regexp pattern")
	}

	return func(event *apiv1.Event) bool {
		return kubeNamePattern.Match([]byte(event.InvolvedObject.Name))
	}, nil
}

func (s *kubeEventService) ListAllKubeEventsByDeploymentSnapshot(ctx context.Context, deployment *models.Deployment, deploymentSnapshot **models.DeploymentSnapshot) ([]apiv1.Event, error) {
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}

	_, eventLister, err := GetEventInformer(ctx, cluster, DeploymentService.GetKubeNamespace(deployment))
	if err != nil {
		return nil, errors.Wrap(err, "get app pool event informer")
	}

	_events, err := eventLister.List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "list events from app pool event informer")
	}

	filter, err := s.MakeKubeEventFilter(ctx, deployment, deploymentSnapshot)
	if err != nil {
		return nil, err
	}

	events := make([]apiv1.Event, 0, len(_events))
	for _, e := range _events {
		if e == nil {
			continue
		}
		if !filter(e) {
			continue
		}
		events = append(events, *e)
	}

	return s.FillKubeEventsType(events), nil
}

func (s *kubeEventService) ListKubeEventsByResourceName(ctx context.Context, deployment *models.Deployment, resourceKind, resourceName string) ([]apiv1.Event, error) {
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	client, _, err := ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "get kube cli set")
	}
	namespace := DeploymentService.GetKubeNamespace(deployment)

	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", resourceKind, resourceName))

	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
		FieldSelector: fieldSelector.String(),
	})

	if err != nil {
		return nil, err
	}

	return s.FillKubeEventsType(list.Items), nil
}

func (s *kubeEventService) ListKubePodsEvents(ctx context.Context, deployment *models.Deployment, pods []apiv1.Pod) ([]apiv1.Event, error) {
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	client, _, err := ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "get kube cli set")
	}
	namespace := DeploymentService.GetKubeNamespace(deployment)

	list, err := client.CoreV1().Events(namespace).List(ctx, consts.KubeListEverything)
	if err != nil {
		return nil, err
	}

	events := s.filterKubeEventsByPodsUID(list.Items, pods)

	return s.FillKubeEventsType(events), nil
}

func (s *kubeEventService) GetKubePodsEventsMapping(events []apiv1.Event, pods []apiv1.Pod) map[types.UID][]apiv1.Event {
	events = s.filterKubeEventsByPodsUID(events, pods)
	res := make(map[types.UID][]apiv1.Event)
	for _, event := range events {
		events_, ok := res[event.InvolvedObject.UID]
		if !ok {
			events_ = make([]apiv1.Event, 0)
		}
		events_ = append(events_, event)
		res[event.InvolvedObject.UID] = events_
	}
	return res
}

func (s *kubeEventService) GetKubePodsWarningEventsMapping(events []apiv1.Event, pods []apiv1.Pod) map[types.UID][]apiv1.Event {
	// Filter out only warning events
	events = s.FilterWarningKubeEvents(events)
	failedPods := make([]apiv1.Pod, 0)

	// Filter out ready and successful pods
	for _, pod := range pods {
		if !s.isKubePodReadyOrSucceeded(pod) {
			failedPods = append(failedPods, pod)
		}
	}

	// Filter events by failed pods UID
	events = s.filterKubeEventsByPodsUID(events, failedPods)
	events = s.removeDuplicateKubeEvents(events)

	return s.GetKubePodsEventsMapping(events, pods)
}

func (s *kubeEventService) ListKubeClusterNodeEvents(ctx context.Context, cluster *models.Cluster, nodeName string) ([]apiv1.Event, error) {
	eventInformer, _, err := GetNodeEventInformer(ctx, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "get app pool event informer")
	}

	_events, err := eventInformer.Lister().List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "list events from kube cluster node event informer")
	}

	events := make([]apiv1.Event, 0, len(_events))
	for _, e := range _events {
		if e == nil {
			continue
		}
		if e.InvolvedObject.Name == nodeName {
			events = append(events, *e)
		}
	}

	return s.FillKubeEventsType(events), nil
}
