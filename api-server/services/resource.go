package services

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/pkg/errors"
)

type resourceService struct{}

var ResourceService = resourceService{}

func (m *resourceService) Get(ctx context.Context, resourceType models.ResourceType, resourceId uint) (models.IResource, error) {
	switch resourceType {
	case models.ResourceTypeUser:
		user, err := UserService.Get(ctx, resourceId)
		return user, err
	case models.ResourceTypeOrganization:
		org, err := OrganizationService.Get(ctx, resourceId)
		return org, err
	case models.ResourceTypeCluster:
		cluster, err := ClusterService.Get(ctx, resourceId)
		return cluster, err
	case models.ResourceTypeBundle:
		bundle, err := BundleService.Get(ctx, resourceId)
		return bundle, err
	case models.ResourceTypeBundleVersion:
		bundleVersion, err := BundleVersionService.Get(ctx, resourceId)
		return bundleVersion, err
	default:
		return nil, errors.Errorf("cannot recognize this resource type: %s", resourceType)
	}
}
