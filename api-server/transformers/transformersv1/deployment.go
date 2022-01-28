package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"
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
	status_ := modelschemas.DeploymentRevisionStatusActive
	deploymentIds := make([]uint, 0, len(deployments))

	for _, deployment := range deployments {
		deploymentIds = append(deploymentIds, deployment.ID)
	}

	deploymentRevisions, _, err := services.DeploymentRevisionService.List(ctx, services.ListDeploymentRevisionOption{
		DeploymentIds: utils.UintSlicePtr(deploymentIds),
		Status:        &status_,
	})
	if err != nil {
		return nil, err
	}

	deploymentIdToDeploymentRevisionUid := make(map[uint]string)
	for _, deploymentRevision := range deploymentRevisions {
		deploymentIdToDeploymentRevisionUid[deploymentRevision.DeploymentId] = deploymentRevision.Uid
	}

	deploymentRevisionSchemas, err := ToDeploymentRevisionSchemas(ctx, deploymentRevisions)
	if err != nil {
		return nil, err
	}

	deploymentRevisionSchemasMap := make(map[string]*schemasv1.DeploymentRevisionSchema)
	for _, deploymentRevisionSchema := range deploymentRevisionSchemas {
		deploymentRevisionSchemasMap[deploymentRevisionSchema.Uid] = deploymentRevisionSchema
	}

	resourceSchemaMap, err := ToResourceSchemasMap(ctx, deployments)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}

	res := make([]*schemasv1.DeploymentSchema, 0, len(deployments))
	for _, deployment := range deployments {
		deploymentRevisionUid := deploymentIdToDeploymentRevisionUid[deployment.ID]
		deploymentRevisionSchema := deploymentRevisionSchemasMap[deploymentRevisionUid]
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
		resourceSchema, ok := resourceSchemaMap[deployment.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for deployment %s", deployment.GetUid())
		}
		res = append(res, &schemasv1.DeploymentSchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Cluster:        clusterSchema,
			Status:         deployment.Status,
			LatestRevision: deploymentRevisionSchema,
			URLs:           urls,
			KubeNamespace:  deployment.KubeNamespace,
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
