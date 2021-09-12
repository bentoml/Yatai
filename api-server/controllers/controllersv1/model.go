package controllersv1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type modelController struct {
	baseController
}

var ModelController = modelController{}

type GetModelSchema struct {
	GetOrganizationSchema
	ModelName string `path:"modelName"`
}

func (s *GetModelSchema) GetModel(ctx context.Context) (*models.Model, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	model, err := services.ModelService.GetByName(ctx, organization.ID, s.ModelName)
	if err != nil {
		return nil, errors.Wrapf(err, "get model %s", s.ModelName)
	}
	return model, nil
}

func (c *modelController) canView(ctx context.Context, model *models.Model) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canView(ctx, organization)
}

func (c *modelController) canUpdate(ctx context.Context, model *models.Model) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canUpdate(ctx, organization)
}

func (c *modelController) canOperate(ctx context.Context, model *models.Model) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, model)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canOperate(ctx, organization)
}

type CreateModelSchema struct {
	schemasv1.CreateModelSchema
	GetOrganizationSchema
}

func (c *modelController) Create(ctx *gin.Context, schema *CreateModelSchema) (*schemasv1.ModelSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = OrganizationController.canUpdate(ctx, organization); err != nil {
		return nil, err
	}

	model, err := services.ModelService.Create(ctx, services.CreateModelOption{
		OrganizationId: organization.ID,
		CreatorId:      user.ID,
		Name:           schema.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create model")
	}
	return transformersv1.ToModelSchema(ctx, model)
}

type UpdateModelSchema struct {
	schemasv1.UpdateModelSchema
	GetModelSchema
}

func (c *modelController) Update(ctx *gin.Context, schema *UpdateModelSchema) (*schemasv1.ModelSchema, error) {
	model, err := schema.GetModel(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, model); err != nil {
		return nil, err
	}
	model, err = services.ModelService.Update(ctx, model, services.UpdateModelOption{
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update model")
	}
	return transformersv1.ToModelSchema(ctx, model)
}

func (c *modelController) Get(ctx *gin.Context, schema *GetModelSchema) (*schemasv1.ModelSchema, error) {
	model, err := schema.GetModel(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, model); err != nil {
		return nil, err
	}
	return transformersv1.ToModelSchema(ctx, model)
}

type ListModelSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
}

func (c *modelController) List(ctx *gin.Context, schema *ListModelSchema) (*schemasv1.ModelListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	models, total, err := services.ModelService.List(ctx, services.ListModelOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: utils.UintPtr(organization.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list models")
	}

	modelSchemas, err := transformersv1.ToModelSchemas(ctx, models)
	return &schemasv1.ModelListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: modelSchemas,
	}, err
}
