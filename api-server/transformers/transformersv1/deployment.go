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
	status_ := modelschemas.DeploymentSnapshotStatusActive
	type_ := modelschemas.DeploymentSnapshotTypeStable

	deploymentSnapshots, _, err := services.DeploymentSnapshotService.List(ctx, services.ListDeploymentSnapshotOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		DeploymentId: deployment.ID,
		Type:         &type_,
		Status:       &status_,
	})
	if err != nil {
		return nil, err
	}

	var latestDeploymentSnapshot *models.DeploymentSnapshot
	if len(deploymentSnapshots) != 0 {
		latestDeploymentSnapshot = deploymentSnapshots[0]
	}

	var latestDeploymentSnapshotSchema **schemasv1.DeploymentSnapshotSchema
	if latestDeploymentSnapshot != nil {
		latestDeploymentSnapshotSchema_, err := ToDeploymentSnapshotSchema(ctx, latestDeploymentSnapshot)
		if err != nil {
			return nil, err
		}
		latestDeploymentSnapshotSchema = &latestDeploymentSnapshotSchema_
	}

	return &schemasv1.DeploymentFullSchema{
		DeploymentSchema: *s,
		LatestSnapshot:   latestDeploymentSnapshotSchema,
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
