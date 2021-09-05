package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToDeploymentSnapshotSchema(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (*schemasv1.DeploymentSnapshotSchema, error) {
	if deploymentSnapshot == nil {
		return nil, nil
	}
	ss, err := ToDeploymentSnapshotSchemas(ctx, []*models.DeploymentSnapshot{deploymentSnapshot})
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSnapshotSchemas")
	}
	return ss[0], nil
}

func ToDeploymentSnapshotSchemas(ctx context.Context, deploymentSnapshots []*models.DeploymentSnapshot) ([]*schemasv1.DeploymentSnapshotSchema, error) {
	res := make([]*schemasv1.DeploymentSnapshotSchema, 0, len(deploymentSnapshots))
	for _, deploymentSnapshot := range deploymentSnapshots {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, deploymentSnapshot)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		bentoVersionFullSchema, err := GetAssociatedBentoVersionFullSchema(ctx, deploymentSnapshot)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBentoVersionFullSchema")
		}
		res = append(res, &schemasv1.DeploymentSnapshotSchema{
			ResourceSchema: ToResourceSchema(deploymentSnapshot),
			Creator:        creatorSchema,
			Type:           deploymentSnapshot.Type,
			Status:         deploymentSnapshot.Status,
			BentoVersion:   bentoVersionFullSchema,
			CanaryRules:    deploymentSnapshot.CanaryRules,
			Config:         deploymentSnapshot.Config,
		})
	}
	return res, nil
}

type IDeploymentSnapshotAssociate interface {
	services.IDeploymentSnapshotAssociate
	models.IResource
}

func GetAssociatedDeploymentSnapshotSchema(ctx context.Context, associate IDeploymentSnapshotAssociate) (*schemasv1.DeploymentSnapshotSchema, error) {
	user, err := services.DeploymentSnapshotService.GetAssociatedDeploymentSnapshot(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	userSchema, err := ToDeploymentSnapshotSchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSnapshotSchema")
	}
	return userSchema, nil
}
