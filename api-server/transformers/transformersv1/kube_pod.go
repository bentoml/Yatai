package transformersv1

import (
	"context"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
)

func ToKubePodSchema(ctx context.Context, clusterId uint, pod *models.KubePodWithStatus) (v *schemasv1.KubePodSchema, err error) {
	vs, err := ToKubePodSchemas(ctx, clusterId, []*models.KubePodWithStatus{pod})
	if err != nil {
		return nil, err
	}
	return vs[0], nil
}

func ToKubePodSchemas(ctx context.Context, clusterId uint, pods []*models.KubePodWithStatus) (vs []*schemasv1.KubePodSchema, err error) {
	sort.SliceStable(pods, func(i, j int) bool {
		iName := pods[i].Pod.Name
		jName := pods[j].Pod.Name

		return strings.Compare(iName, jName) >= 0
	})

	sort.SliceStable(pods, func(i, j int) bool {
		it := pods[i].Pod.Status.StartTime
		jt := pods[j].Pod.Status.StartTime

		if it == nil {
			return false
		}

		if jt == nil {
			return true
		}

		return it.Before(jt)
	})

	sort.SliceStable(pods, func(i, j int) bool {
		return pods[i].Pod.Labels[consts.KubeLabelYataiDeploymentTargetType] == string(modelschemas.DeploymentTargetTypeStable)
	})

	var deployment *models.Deployment

	for _, p := range pods {
		deploymentName, ok := p.Pod.Labels[consts.KubeLabelYataiDeployment]
		if ok {
			namespace := p.Pod.Namespace
			deployment, err = services.DeploymentService.GetByName(ctx, clusterId, namespace, deploymentName)
			if err != nil {
				return
			}
			break
		}
	}

	deploymentTargetSchemasMap := make(map[modelschemas.DeploymentTargetType]*schemasv1.DeploymentTargetSchema, 2)

	if deployment != nil {
		status := modelschemas.DeploymentRevisionStatusActive
		var deploymentRevisions []*models.DeploymentRevision
		deploymentRevisions, _, err = services.DeploymentRevisionService.List(ctx, services.ListDeploymentRevisionOption{
			BaseListOption: services.BaseListOption{
				Start: utils.UintPtr(0),
				Count: utils.UintPtr(10),
			},
			DeploymentId: utils.UintPtr(deployment.ID),
			Status:       &status,
		})
		if err != nil {
			return
		}
		var deploymentRevisionSchemas []*schemasv1.DeploymentRevisionSchema
		deploymentRevisionSchemas, err = ToDeploymentRevisionSchemas(ctx, deploymentRevisions)
		if err != nil {
			return
		}
		for _, deploymentRevisionSchema := range deploymentRevisionSchemas {
			for _, deploymentTargetSchema := range deploymentRevisionSchema.Targets {
				deploymentTargetSchemasMap[deploymentTargetSchema.Type] = deploymentTargetSchema
			}
		}
	}

	for _, p := range pods {
		var statusReady bool
		for _, c := range p.Pod.Status.Conditions {
			if c.Type == apiv1.PodReady {
				statusReady = c.Status == apiv1.ConditionTrue
			}
		}
		deploymentTargetType, deploymentTargetTypeExists := p.Pod.Labels[consts.KubeLabelYataiDeploymentTargetType]
		var deploymentTargetSchema *schemasv1.DeploymentTargetSchema
		if deploymentTargetTypeExists {
			deploymentTargetSchema = deploymentTargetSchemasMap[modelschemas.DeploymentTargetType(deploymentTargetType)]
		}
		isCanary := deploymentTargetTypeExists && deploymentTargetType == string(modelschemas.DeploymentTargetTypeCanary)
		status := schemasv1.KubePodStatusSchema{
			Phase:     p.Pod.Status.Phase,
			Ready:     statusReady,
			StartTime: p.Pod.Status.StartTime,
			IsOld:     false,
			IsCanary:  isCanary,
			HostIp:    p.Pod.Status.HostIP,
		}
		var runnerName *string
		runnerName_, runnerNameExists := p.Pod.Labels[consts.KubeLabelYataiBentoRunner]
		if runnerNameExists {
			runnerName = &runnerName_
		}
		vs = append(vs, &schemasv1.KubePodSchema{
			Name:             p.Pod.Name,
			Namespace:        p.Pod.Namespace,
			NodeName:         p.Pod.Spec.NodeName,
			RunnerName:       runnerName,
			Status:           status,
			RawStatus:        p.Pod.Status,
			PodStatus:        p.Status,
			Warnings:         p.Warnings,
			DeploymentTarget: deploymentTargetSchema,
		})
	}
	return
}
