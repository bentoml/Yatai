package controllersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type organizationController struct {
	baseController
}

var OrganizationController = organizationController{}

type GetOrganizationSchema struct {
	OrgName string `path:"orgName"`
}

func (s *GetOrganizationSchema) GetOrganization(ctx context.Context) (*models.Organization, error) {
	organization, err := services.OrganizationService.GetByName(ctx, s.OrgName)
	if err != nil {
		return nil, errors.Wrapf(err, "get organization %s", s.OrgName)
	}
	return organization, nil
}

func (c *organizationController) canView(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanView(ctx, &services.OrganizationMemberService, user.ID, organization.ID)
}

func (c *organizationController) canUpdate(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanUpdate(ctx, &services.OrganizationMemberService, user.ID, organization.ID)
}

func (c *organizationController) canOperate(ctx context.Context, organization *models.Organization) error {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}
	return services.MemberService.CanOperate(ctx, &services.OrganizationMemberService, user.ID, organization.ID)
}

func (c *organizationController) Create(ctx *gin.Context, schema *schemasv1.CreateOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := services.OrganizationService.Create(ctx, services.CreateOrganizationOption{
		CreatorId:   user.ID,
		Name:        schema.Name,
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create organization")
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

type UpdateOrganizationSchema struct {
	schemasv1.UpdateOrganizationSchema
	GetOrganizationSchema
}

func (c *organizationController) Update(ctx *gin.Context, schema *UpdateOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, organization); err != nil {
		return nil, err
	}
	organization, err = services.OrganizationService.Update(ctx, organization, services.UpdateOrganizationOption{
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update organization")
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

func (c *organizationController) Get(ctx *gin.Context, schema *GetOrganizationSchema) (*schemasv1.OrganizationFullSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, organization); err != nil {
		return nil, err
	}
	return transformersv1.ToOrganizationFullSchema(ctx, organization)
}

func (c *organizationController) List(ctx *gin.Context, schema *schemasv1.ListQuerySchema) (*schemasv1.OrganizationListSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	organizations, total, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{BaseListOption: services.BaseListOption{
		Start:  utils.UintPtr(schema.Start),
		Count:  utils.UintPtr(schema.Count),
		Search: schema.Search,
	},
		VisitorId: utils.UintPtr(user.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list organizations")
	}

	organizationSchemas, err := transformersv1.ToOrganizationSchemas(ctx, organizations)
	return &schemasv1.OrganizationListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: organizationSchemas,
	}, err
}
