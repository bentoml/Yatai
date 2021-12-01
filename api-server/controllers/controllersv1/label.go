package controllersv1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type labelController struct {
	baseController
}

var LabelController = labelController{}

type GetLabelSchema struct {
	GetOrganizationSchema
	Key          string                    `path:"key"`
	ResourceType modelschemas.ResourceType `query:"resource_type"`
	ResourceUid  string                    `query:"resource_uid"`
}

func (s *GetLabelSchema) GetLabel(ctx context.Context) (*models.Label, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	resource, err := services.ResourceService.GetByUid(ctx, s.ResourceType, s.ResourceUid)
	if err != nil {
		return nil, errors.Wrap(err, "get resource")
	}
	label, err := services.LabelService.GetByKey(ctx, services.GetLabelByKeyOption{
		OrganizationId: organization.ID,
		ResourceType:   resource.GetResourceType(),
		ResourceId:     resource.GetId(),
		Key:            s.Key,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get label %s", s.Key)
	}
	return label, nil
}

func (c *labelController) Get(ctx *gin.Context, schema *GetLabelSchema) (*schemasv1.LabelSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = OrganizationController.canView(ctx, org); err != nil {
		return nil, err
	}
	label, err := schema.GetLabel(ctx)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToLabelSchema(ctx, label)
}

type ListLabelSchema struct {
	schemasv1.ListQuerySchema
	GetOrganizationSchema
	ResourceType modelschemas.ResourceType
}

func (c *labelController) List(ctx *gin.Context, schema *ListLabelSchema) (*schemasv1.LabelListSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, org); err != nil {
		return nil, err
	}

	labels, total, err := services.LabelService.List(ctx, services.ListLabelOption{
		OrganizationId: utils.UintPtr(org.ID),
		ResourceType:   schema.ResourceType.Ptr(),
	})
	if err != nil {
		return nil, err
	}
	schemas, err := transformersv1.ToLabelSchemas(ctx, labels)
	if err != nil {
		return nil, err
	}
	return &schemasv1.LabelListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: schemas,
	}, nil
}
