package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToDeploymentTargetSchema(ctx context.Context, deploymentTarget *models.DeploymentTarget) (*schemasv1.DeploymentTargetSchema, error) {
	if deploymentTarget == nil {
		return nil, nil
	}
	ss, err := ToDeploymentTargetSchemas(ctx, []*models.DeploymentTarget{deploymentTarget})
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentTargetSchemas")
	}
	return ss[0], nil
}

func ToDeploymentTargetSchemas(ctx context.Context, deploymentTargets []*models.DeploymentTarget) ([]*schemasv1.DeploymentTargetSchema, error) {
	res := make([]*schemasv1.DeploymentTargetSchema, 0, len(deploymentTargets))
	for _, deploymentTarget := range deploymentTargets {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, deploymentTarget)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		bentoVersionFullSchema, err := GetAssociatedBentoVersionFullSchema(ctx, deploymentTarget)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBentoVersionFullSchema")
		}
		res = append(res, &schemasv1.DeploymentTargetSchema{
			ResourceSchema: ToResourceSchema(deploymentTarget),
			DeploymentTargetTypeSchema: schemasv1.DeploymentTargetTypeSchema{
				Type: deploymentTarget.Type,
			},
			Creator:      creatorSchema,
			BentoVersion: bentoVersionFullSchema,
			CanaryRules:  deploymentTarget.CanaryRules,
			Config:       deploymentTarget.Config,
		})
	}
	return res, nil
}
