package services

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	commonconsts "github.com/bentoml/yatai-common/consts"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
)

type imageBuilderService struct{}

var ImageBuilderService = &imageBuilderService{}

func (s *imageBuilderService) ListImageBuilderPods(ctx context.Context, cluster *models.Cluster, kubeLabels map[string]string) ([]*models.KubePodWithStatus, error) {
	_, podLister, err := GetPodInformer(ctx, cluster, commonconsts.KubeNamespaceYataiModelImageBuilder)
	if err != nil {
		return nil, err
	}

	selectorPieces := make([]string, 0, len(kubeLabels))
	for k, v := range kubeLabels {
		selectorPieces = append(selectorPieces, fmt.Sprintf("%s = %s", k, v))
	}

	selector, err := labels.Parse(strings.Join(selectorPieces, ", "))
	if err != nil {
		return nil, err
	}
	pods, err := podLister.List(selector)
	if err != nil {
		return nil, err
	}
	_, eventLister, err := GetEventInformer(ctx, cluster, commonconsts.KubeNamespaceYataiModelImageBuilder)
	if err != nil {
		return nil, err
	}

	events, err := eventLister.List(selector)
	if err != nil {
		return nil, err
	}

	pods_ := make([]apiv1.Pod, 0, len(pods))
	for _, p := range pods {
		pods_ = append(pods_, *p)
	}
	events_ := make([]apiv1.Event, 0, len(pods))
	for _, e := range events {
		events_ = append(events_, *e)
	}

	pods__ := KubePodService.MapKubePodsToKubePodWithStatuses(ctx, pods_, events_)

	res := make([]*models.KubePodWithStatus, 0)
	for _, pod := range pods__ {
		if pod.Status.Status == modelschemas.KubePodActualStatusTerminating {
			continue
		}
		res = append(res, pod)
	}

	return res, nil
}
