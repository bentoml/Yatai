package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToDeploymentSchema(ctx context.Context, deployment *models.Deployment) (*schemasv1.DeploymentSchema, error) {
	if deployment == nil {
		return nil, nil
	}
	ss, err := ToDeploymentSchemas(ctx, []*models.Deployment{deployment})
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSchemas")
	}
	return ss[0], nil
}

func ToDeploymentSchemas(ctx context.Context, deployments []*models.Deployment) ([]*schemasv1.DeploymentSchema, error) {
	res := make([]*schemasv1.DeploymentSchema, 0, len(deployments))
	for _, deployment := range deployments {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, deployment)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		clusterSchema, err := GetAssociatedClusterFullSchema(ctx, deployment)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedClusterSchema")
		}
		urls, err := services.DeploymentService.GetURLs(ctx, deployment)
		if err != nil {
			return nil, errors.Wrap(err, "get deployment urls")
		}
		res = append(res, &schemasv1.DeploymentSchema{
			ResourceSchema: ToResourceSchema(deployment),
			Creator:        creatorSchema,
			Cluster:        clusterSchema,
			Status:         deployment.Status,
			URLs:           urls,
		})
	}
	return res, nil
}

type IDeploymentAssociate interface {
	services.IDeploymentAssociate
	models.IResource
}

func GetAssociatedDeploymentSchema(ctx context.Context, associate IDeploymentAssociate) (*schemasv1.DeploymentSchema, error) {
	user, err := services.DeploymentService.GetAssociatedDeployment(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToDeploymentSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSchema")
	}
	return userSchema, nil
}
