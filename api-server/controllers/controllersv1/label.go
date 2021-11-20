package controllersv1

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type labelController struct {
	baseController
}

var LabelController = labelController{}

type GetLabelSchema struct {
	Id uint `path:"id"`
}

func (s *GetLabelSchema) GetLabel(ctx context.Context) (*models.Label, error) {
	return services.LabelService.Get(ctx, s.Id)
}

func (c *labelController) canView(ctx context.Context, label *models.Label) error {
	// depending on resource type, check that we can view each Deployment/Model/Bento
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canView(ctx, org)
	}
	return nil
}

func (c *labelController) canUpdate(ctx context.Context, label *models.Label) error {
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canUpdate(ctx, org)
	}
	return nil
}

func (c *labelController) canOperate(ctx context.Context, label *models.Label) error {
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canOperate(ctx, org)
	}
	return nil
}

func (c *labelController) Get(ctx *gin.Context, schema *GetLabelSchema) (*schemasv1.LabelSchema, error) {
	label, err := schema.GetLabel(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, label); err != nil {
		return nil, err
	}
	return transformersv1.ToLabelSchema(ctx, label)
}

/*
TODO:
Create()
Update()
List()
*/
