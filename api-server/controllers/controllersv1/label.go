package controllersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type labelController struct {
	baseController
}

var labelController = labelController{}

type GetLabelSchema struct {
	Id string `path:"id"`
}

func (s *GetLabelSchema) GetLabel(ctx context.Context) (*models.Label, error) {
	return services.LabelService.GetByUid(ctx, s.Uid)
}

func (c *labelController) canView(ctx context.Context, label *models.Label) error {
	// depending on resource type, check that we can view each Deployment/Model/Bento
	cluster, err := services.ClusterService.GetAssociatedNullableCluster(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if cluster != nil {
		return ClusterController.canView(ctx, cluster)
	}

	org, err := services.OrganizationService.GetAssociatedNullableOrganization(ctx, label)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canView(ctx, org)
	}

	bento, err := services.BentoService.GetAssociatedBento(ctx, label)
	if err != nil {
		return err
	}
	if bento != nil {
		return BentoController.canView(ctx, bento)
	}

	bento_version, err := services.BentoVersionService(ctx, label)
	if err != nil {
		return err
	}
	if bento_version != nil {
		return BentoVersionController.canView(ctx, bento_version)
	}

	model, err := services.ModelService(ctx, label)
	if err != nil {
		return err
	}
	if model != nil {
		return ModelController.canView(ctx, model)
	}

	model_version, err := services.ModelVersionService(ctx, label)
	if err != nil {
		return err
	}
	if model_version != nil {
		return ModelVersionController.canView(ctx, model_version)
	}

	deployment, err := services.DeploymentService.GetAssociatedNullableDeployment(ctx, label)
	if err != nil {
		return err
	}
	if deployment != nil {
		return DeploymentController.canView(ctx, deployment)
	}
	return nil
}

func (c *labelController) Get(ctx *gin.Context, schema *GetLabelSchema) (*schemasv1.LabelSchema, error) {
	label, err := schema.getLabel(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, label); err != nil {
		return nil, err
	}
	return transformersv1.ToLabelRecordSchema(ctx, label)
}

/*
TODO:
canUpdate()
canOperate()
Create()
Update()
List()
*/
