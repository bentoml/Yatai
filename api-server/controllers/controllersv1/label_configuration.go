package controllersv1

import (
	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type labelConfigurationController struct{}

var LabelConfigurationController = labelConfigurationController{}

type CreateLabelConfigurationSchema struct {
	GetOrganizationSchema
	schemasv1.CreateLabelConfigurationSchema
}

type GetLabelConfigurationSchema struct {
	GetOrganizationSchema
	Key string `path:"key"`
}

func (s *GetLabelConfigurationSchema) GetLabelConfiguration(ctx *gin.Context) (*models.LabelConfiguration, error) {
	org, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	return services.LabelConfigurationService.GetByKey(ctx, org.ID, s.Key)
}

type UpdateLabelConfigurationSchema struct {
	GetLabelConfigurationSchema
	schemasv1.UpdateLabelConfigurationSchema
}

func (c *labelConfigurationController) Create(ctx *gin.Context, schema *CreateLabelConfigurationSchema) (*schemasv1.LabelConfigurationSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	err = OrganizationController.canUpdate(ctx, org)
	if err != nil {
		return nil, err
	}
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	labelConfiguration, err := services.LabelConfigurationService.Create(ctx, services.CreateLabelConfigurationOption{
		OrganizationId: org.ID,
		CreatorId:      currentUser.ID,
		Key:            schema.Key,
		Info:           schema.Info,
	})
	if err != nil {
		return nil, err
	}
	return transformersv1.ToLabelConfigurationSchema(ctx, labelConfiguration)
}

func (c *labelConfigurationController) Get(ctx *gin.Context, schema *GetLabelConfigurationSchema) (*schemasv1.LabelConfigurationSchema, error) {
	labelConfiguration, err := schema.GetLabelConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToLabelConfigurationSchema(ctx, labelConfiguration)
}

func (c *labelConfigurationController) Update(ctx *gin.Context, schema *UpdateLabelConfigurationSchema) (*schemasv1.LabelConfigurationSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	err = OrganizationController.canUpdate(ctx, org)
	if err != nil {
		return nil, err
	}
	labelConfiguration, err := schema.GetLabelConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	labelConfiguration, err = services.LabelConfigurationService.Update(ctx, labelConfiguration, services.UpdateLabelConfigurationOption{
		Info: schema.Info,
	})
	if err != nil {
		return nil, err
	}
	return transformersv1.ToLabelConfigurationSchema(ctx, labelConfiguration)
}

type ListLabelConfigurationSchema struct {
	GetOrganizationSchema
	schemasv1.ListQuerySchema
}

func (c *labelConfigurationController) List(ctx *gin.Context, schema *ListLabelConfigurationSchema) (*schemasv1.LabelConfigurationListSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	err = OrganizationController.canView(ctx, org)
	if err != nil {
		return nil, err
	}
	labelConfigurations, total, err := services.LabelConfigurationService.List(ctx, services.ListLabelConfigurationOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: org.ID,
	})
	if err != nil {
		return nil, err
	}
	schemas, err := transformersv1.ToLabelConfigurationSchemas(ctx, labelConfigurations)
	if err != nil {
		return nil, err
	}
	return &schemasv1.LabelConfigurationListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Start: schema.Start,
			Count: schema.Count,
			Total: total,
		},
		Items: schemas,
	}, nil
}

func (c *labelConfigurationController) Delete(ctx *gin.Context, schema *GetLabelConfigurationSchema) (*schemasv1.LabelConfigurationSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	err = OrganizationController.canUpdate(ctx, org)
	if err != nil {
		return nil, err
	}
	labelConfiguration, err := schema.GetLabelConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	labelConfiguration, err = services.LabelConfigurationService.Delete(ctx, labelConfiguration)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToLabelConfigurationSchema(ctx, labelConfiguration)
}
