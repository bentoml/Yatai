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
	GetOrganizationSchema
	Key string `path:"key"`
}

func (s *GetLabelSchema) GetLabel(ctx context.Context) (*models.Label, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	label, err := services.LabelService.GetByKey(ctx, organization.ID, s.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "get label %s", s.Key)
	}
	return label, nil
}

func (c *labelController) canView(ctx context.Context, label *models.Label) error {
	// depending on resource type, check that we can view each Deployment/Model/Bento
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	return OrganizationController.canView(ctx, org)
}

func (c *labelController) canUpdate(ctx context.Context, label *models.Label) error {
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	return OrganizationController.canUpdate(ctx, org)
}

func (c *labelController) canOperate(ctx context.Context, label *models.Label) error {
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, label)
	if err != nil {
		return err
	}
	return OrganizationController.canOperate(ctx, org)
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

type CreateLabelSchema struct {
	schemasv1.CreateLabelSchema
	GetOrganizationSchema
}

func (c *labelController) Create(ctx *gin.Context, schema *CreateLabelSchema) (*schemasv1.LabelSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = OrganizationController.canUpdate(ctx, organization); err != nil {
		return nil, err
	}
	if err = LabelController.canUpdate(ctx, label); err != nil {
		return nil, err
	}
	label, err := services.LabelService.Create(ctx, services.CreateLabelOption{
		OrganizationId: org.ID,
		CreatorId:      user.ID,
		Resource:       schema.Resource,
		Key:            schema.Key,
		Value:          schema.Value,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create label")
	}
	return c.doUpdate(ctx, schema.UpdateLabelSchema, org, label)
}

type UpdateLabelSchema struct {
	schemasv1.UpdateLabelSchema
	GetOrganizationSchema
}

func (c *labelController) Update(ctx *gin.Context, schema *UpdateLabelSchema) (*schemasv1.LabelSchema, error) {
	label, err := schema.GetLabel(ctx)
	if err != nil {
		return nil, err
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, label); err != nil {
		return nil, err
	}
	label, err = services.LabelService.Update(ctx, label, services.UpdateLabelOption{
		Value: schema.Value,
	})
	if err != nil {
		return nil, errors.wrap(err, "update label")
	}
	return c.doUpdate(ctx, schema.UpdateLabelSchema, org, label)
}

func (c *labelController) doUpdate(ctx *gin.Context, schema schemasv1.UpdateLabelSchema, org *models.Organization, label *models.Label) (*schemasv1.LabelSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// TODO?
}
