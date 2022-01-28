package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToDeploymentRevisionSchema(ctx context.Context, deploymentRevision *models.DeploymentRevision) (*schemasv1.DeploymentRevisionSchema, error) {
	if deploymentRevision == nil {
		return nil, nil
	}
	ss, err := ToDeploymentRevisionSchemas(ctx, []*models.DeploymentRevision{deploymentRevision})
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentRevisionSchemas")
	}
	return ss[0], nil
}

func ToDeploymentRevisionSchemas(ctx context.Context, deploymentRevisions []*models.DeploymentRevision) ([]*schemasv1.DeploymentRevisionSchema, error) {
	deploymentRevisionIds := make([]uint, 0, len(deploymentRevisions))
	for _, deploymentRevision := range deploymentRevisions {
		deploymentRevisionIds = append(deploymentRevisionIds, deploymentRevision.ID)
	}
	deploymentTargets, _, err := services.DeploymentTargetService.List(ctx, services.ListDeploymentTargetOption{
		DeploymentRevisionIds: utils.UintSlicePtr(deploymentRevisionIds),
	})
	if err != nil {
		return nil, err
	}
	deploymentTargetsMapping := make(map[uint][]*models.DeploymentTarget)
	for _, deploymentTarget := range deploymentTargets {
		deploymentTargets, ok := deploymentTargetsMapping[deploymentTarget.DeploymentRevisionId]
		if !ok {
			deploymentTargets = make([]*models.DeploymentTarget, 0)
		}
		deploymentTargets = append(deploymentTargets, deploymentTarget)
		deploymentTargetsMapping[deploymentTarget.DeploymentRevisionId] = deploymentTargets
	}
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, deploymentRevisions)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	res := make([]*schemasv1.DeploymentRevisionSchema, 0, len(deploymentRevisions))
	for _, deploymentRevision := range deploymentRevisions {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, deploymentRevision)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		deploymentTargets := deploymentTargetsMapping[deploymentRevision.ID]
		deploymentTargetSchemas, err := ToDeploymentTargetSchemas(ctx, deploymentTargets)
		if err != nil {
			return nil, errors.Wrap(err, "ToDeploymentTargetSchemas")
		}
		resourceSchema, ok := resourceSchemasMap[deploymentRevision.GetUid()]
		if !ok {
			return nil, errors.Errorf("resourceSchema not found for deploymentRevision %s", deploymentRevision.GetUid())
		}
		res = append(res, &schemasv1.DeploymentRevisionSchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Status:         deploymentRevision.Status,
			Targets:        deploymentTargetSchemas,
		})
	}
	return res, nil
}

type IDeploymentRevisionAssociate interface {
	services.IDeploymentRevisionAssociate
	models.IResource
}

func GetAssociatedDeploymentRevisionSchema(ctx context.Context, associate IDeploymentRevisionAssociate) (*schemasv1.DeploymentRevisionSchema, error) {
	user, err := services.DeploymentRevisionService.GetAssociatedDeploymentRevision(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToDeploymentRevisionSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentRevisionSchema")
	}
	return userSchema, nil
}
