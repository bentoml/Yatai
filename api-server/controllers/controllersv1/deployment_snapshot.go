package controllersv1

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type deploymentSnapshotController struct {
	baseController
}

var DeploymentSnapshotController = deploymentSnapshotController{}

type ListDeploymentSnapshotSchema struct {
	schemasv1.ListQuerySchema
	GetDeploymentSchema
}

func (c *deploymentSnapshotController) List(ctx *gin.Context, schema *ListDeploymentSnapshotSchema) (*schemasv1.DeploymentSnapshotListSchema, error) {
	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}

	if err = DeploymentController.canView(ctx, deployment); err != nil {
		return nil, err
	}

	deploymentSnapshots, total, err := services.DeploymentSnapshotService.List(ctx, services.ListDeploymentSnapshotOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		DeploymentId: deployment.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list deploymentSnapshots")
	}

	deploymentSchemas, err := transformersv1.ToDeploymentSnapshotSchemas(ctx, deploymentSnapshots)
	return &schemasv1.DeploymentSnapshotListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: deploymentSchemas,
	}, err
}
