package services

import (
	"context"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
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
	case modelschemas.ResourceTypeBento:
		bento, err := BentoService.Get(ctx, resourceId)
		return bento, err
	case modelschemas.ResourceTypeBentoVersion:
		bentoVersion, err := BentoVersionService.Get(ctx, resourceId)
		return bentoVersion, err
	case modelschemas.ResourceTypeDeployment:
		deployment, err := DeploymentService.Get(ctx, resourceId)
		return deployment, err
	case modelschemas.ResourceTypeDeploymentRevision:
		deploymentRevision, err := DeploymentRevisionService.Get(ctx, resourceId)
		return deploymentRevision, err
	case modelschemas.ResourceTypeTerminalRecord:
		terminalRecord, err := TerminalRecordService.Get(ctx, resourceId)
		return terminalRecord, err
	case modelschemas.ResourceTypeModel:
		model, err := ModelService.Get(ctx, resourceId)
		return model, err
	case modelschemas.ResourceTypeModelVersion:
		modelVersion, err := ModelVersionService.Get(ctx, resourceId)
		return modelVersion, err
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
	case modelschemas.ResourceTypeBento:
		bento, err := BentoService.GetByUid(ctx, resourceUid)
		return bento, err
	case modelschemas.ResourceTypeBentoVersion:
		bentoVersion, err := BentoVersionService.GetByUid(ctx, resourceUid)
		return bentoVersion, err
	case modelschemas.ResourceTypeDeployment:
		deployment, err := DeploymentService.GetByUid(ctx, resourceUid)
		return deployment, err
	case modelschemas.ResourceTypeDeploymentRevision:
		deploymentRevision, err := DeploymentRevisionService.GetByUid(ctx, resourceUid)
		return deploymentRevision, err
	case modelschemas.ResourceTypeTerminalRecord:
		terminalRecord, err := TerminalRecordService.GetByUid(ctx, resourceUid)
		return terminalRecord, err
	case modelschemas.ResourceTypeModel:
		model, err := ModelService.GetByUid(ctx, resourceUid)
		return model, err
	case modelschemas.ResourceTypeModelVersion:
		modelVersion, err := ModelVersionService.GetByUid(ctx, resourceUid)
		return modelVersion, err
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
