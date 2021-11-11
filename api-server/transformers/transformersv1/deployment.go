package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToDeploymentFullSchema(ctx context.Context, deployment *models.Deployment) (*schemasv1.DeploymentFullSchema, error) {
	s, err := ToDeploymentSchema(ctx, deployment)
	if err != nil {
		return nil, err
	}
	status_ := modelschemas.DeploymentRevisionStatusActive

	deploymentRevisions, _, err := services.DeploymentRevisionService.List(ctx, services.ListDeploymentRevisionOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		DeploymentId: deployment.ID,
		Status:       &status_,
	})
	if err != nil {
		return nil, err
	}

	var latestDeploymentRevision *models.DeploymentRevision
	if len(deploymentRevisions) != 0 {
		latestDeploymentRevision = deploymentRevisions[0]
	}

	var latestDeploymentRevisionSchema *schemasv1.DeploymentRevisionSchema
	if latestDeploymentRevision != nil {
		latestDeploymentRevisionSchema_, err := ToDeploymentRevisionSchema(ctx, latestDeploymentRevision)
		if err != nil {
			return nil, err
		}
		latestDeploymentRevisionSchema = latestDeploymentRevisionSchema_
	}

	return &schemasv1.DeploymentFullSchema{
		DeploymentSchema: *s,
		LatestRevision:   latestDeploymentRevisionSchema,
	}, nil
}

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
	deployment, err := services.DeploymentService.GetAssociatedDeployment(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	deploymentSchema, err := ToDeploymentSchema(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSchema")
	}
	return deploymentSchema, nil
}

type INullableDeploymentAssociate interface {
	services.INullableDeploymentAssociate
	models.IResource
}

func GetAssociatedNullableDeploymentSchema(ctx context.Context, associate INullableDeploymentAssociate) (*schemasv1.DeploymentSchema, error) {
	deployment, err := services.DeploymentService.GetAssociatedNullableDeployment(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	if deployment == nil {
		return nil, nil
	}
	deploymentSchema, err := ToDeploymentSchema(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "ToNullableDeploymentSchema")
	}
	return deploymentSchema, nil
}
