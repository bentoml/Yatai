package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type resourceService struct{}

var ResourceService = resourceService{}

func (m *resourceService) Get(ctx context.Context, resourceType modelschemas.ResourceType, resourceId uint) (models.IResource, error) {
	switch resourceType {
	case modelschemas.ResourceTypeUser:
		user, err := UserService.Get(ctx, resourceId)
		return user, err
	case modelschemas.ResourceTypeOrganization:
		org, err := OrganizationService.Get(ctx, resourceId)
		return org, err
	case modelschemas.ResourceTypeCluster:
		cluster, err := ClusterService.Get(ctx, resourceId)
		return cluster, err
	case modelschemas.ResourceTypeBentoRepository:
		bentoRepository, err := BentoRepositoryService.Get(ctx, resourceId)
		return bentoRepository, err
	case modelschemas.ResourceTypeBento:
		bento, err := BentoService.Get(ctx, resourceId)
		return bento, err
	case modelschemas.ResourceTypeDeployment:
		deployment, err := DeploymentService.Get(ctx, resourceId)
		return deployment, err
	case modelschemas.ResourceTypeDeploymentRevision:
		deploymentRevision, err := DeploymentRevisionService.Get(ctx, resourceId)
		return deploymentRevision, err
	case modelschemas.ResourceTypeTerminalRecord:
		terminalRecord, err := TerminalRecordService.Get(ctx, resourceId)
		return terminalRecord, err
	case modelschemas.ResourceTypeModelRepository:
		modelRepository, err := ModelRepositoryService.Get(ctx, resourceId)
		return modelRepository, err
	case modelschemas.ResourceTypeModel:
		model, err := ModelService.Get(ctx, resourceId)
		return model, err
	case modelschemas.ResourceTypeApiToken:
		apiToken, err := ApiTokenService.Get(ctx, resourceId)
		return apiToken, err
	case modelschemas.ResourceTypeLabel:
		label, err := LabelService.Get(ctx, resourceId)
		return label, err
	default:
		return nil, errors.Errorf("cannot recognize this resource type: %s", resourceType)
	}
}

func (m *resourceService) List(ctx context.Context, resourceType modelschemas.ResourceType, resourceIds []uint) (interface{}, error) {
	switch resourceType {
	case modelschemas.ResourceTypeUser:
		users, err := UserService.ListByIds(ctx, resourceIds)
		return users, err
	case modelschemas.ResourceTypeOrganization:
		orgs, _, err := OrganizationService.List(ctx, ListOrganizationOption{
			Ids: &resourceIds,
		})
		return orgs, err
	case modelschemas.ResourceTypeCluster:
		clusters, _, err := ClusterService.List(ctx, ListClusterOption{
			Ids: &resourceIds,
		})
		return clusters, err
	case modelschemas.ResourceTypeBentoRepository:
		bentoRepositories, _, err := BentoRepositoryService.List(ctx, ListBentoRepositoryOption{
			Ids: &resourceIds,
		})
		return bentoRepositories, err
	case modelschemas.ResourceTypeBento:
		bentos, _, err := BentoService.List(ctx, ListBentoOption{
			Ids: &resourceIds,
		})
		return bentos, err
	case modelschemas.ResourceTypeDeployment:
		deployments, _, err := DeploymentService.List(ctx, ListDeploymentOption{
			Ids: &resourceIds,
		})
		return deployments, err
	case modelschemas.ResourceTypeDeploymentRevision:
		deploymentRevisions, _, err := DeploymentRevisionService.List(ctx, ListDeploymentRevisionOption{
			Ids: &resourceIds,
		})
		return deploymentRevisions, err
	case modelschemas.ResourceTypeTerminalRecord:
		terminalRecords, _, err := TerminalRecordService.List(ctx, ListTerminalRecordOption{
			Ids: &resourceIds,
		})
		return terminalRecords, err
	case modelschemas.ResourceTypeModelRepository:
		modelRepositories, _, err := ModelRepositoryService.List(ctx, ListModelRepositoryOption{
			Ids: &resourceIds,
		})
		return modelRepositories, err
	case modelschemas.ResourceTypeModel:
		models, _, err := ModelService.List(ctx, ListModelOption{
			Ids: &resourceIds,
		})
		return models, err
	case modelschemas.ResourceTypeApiToken:
		apiTokens, _, err := ApiTokenService.List(ctx, ListApiTokenOption{
			Ids: &resourceIds,
		})
		return apiTokens, err
	case modelschemas.ResourceTypeLabel:
		labels, _, err := LabelService.List(ctx, ListLabelOption{
			Ids: &resourceIds,
		})
		return labels, err
	default:
		return nil, errors.Errorf("cannot recognize this resource type: %s", resourceType)
	}
}

func (m *resourceService) GetByUid(ctx context.Context, resourceType modelschemas.ResourceType, resourceUid string) (models.IResource, error) {
	switch resourceType {
	case modelschemas.ResourceTypeUser:
		user, err := UserService.GetByUid(ctx, resourceUid)
		return user, err
	case modelschemas.ResourceTypeOrganization:
		org, err := OrganizationService.GetByUid(ctx, resourceUid)
		return org, err
	case modelschemas.ResourceTypeCluster:
		cluster, err := ClusterService.GetByUid(ctx, resourceUid)
		return cluster, err
	case modelschemas.ResourceTypeBentoRepository:
		bentoRepository, err := BentoRepositoryService.GetByUid(ctx, resourceUid)
		return bentoRepository, err
	case modelschemas.ResourceTypeBento:
		bento, err := BentoService.GetByUid(ctx, resourceUid)
		return bento, err
	case modelschemas.ResourceTypeDeployment:
		deployment, err := DeploymentService.GetByUid(ctx, resourceUid)
		return deployment, err
	case modelschemas.ResourceTypeDeploymentRevision:
		deploymentRevision, err := DeploymentRevisionService.GetByUid(ctx, resourceUid)
		return deploymentRevision, err
	case modelschemas.ResourceTypeTerminalRecord:
		terminalRecord, err := TerminalRecordService.GetByUid(ctx, resourceUid)
		return terminalRecord, err
	case modelschemas.ResourceTypeModelRepository:
		modelRepository, err := ModelRepositoryService.GetByUid(ctx, resourceUid)
		return modelRepository, err
	case modelschemas.ResourceTypeModel:
		model, err := ModelService.GetByUid(ctx, resourceUid)
		return model, err
	case modelschemas.ResourceTypeApiToken:
		apiToken, err := ApiTokenService.GetByUid(ctx, resourceUid)
		return apiToken, err
	case modelschemas.ResourceTypeLabel:
		label, err := LabelService.GetByUid(ctx, resourceUid)
		return label, err
	default:
		return nil, errors.Errorf("cannot recognize this resource type: %s", resourceType)
	}
}
