package controllersv1

import (
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type organizationMemberController struct {
	organizationController
}

var OrganizationMemberController = organizationMemberController{}

type CreateOrganizationMemberSchema struct {
	schemasv1.CreateOrganizationMemberSchema
	GetOrganizationSchema
}

func (c *organizationMemberController) Create(ctx *gin.Context, schema *CreateOrganizationMemberSchema) (*schemasv1.OrganizationMemberSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current user")
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, org); err != nil {
		return nil, err
	}
	organizationMember, err := services.OrganizationMemberService.Create(ctx, user.ID, services.CreateOrganizationMemberOption{
		CreatorId:      schema.UserId,
		OrganizationId: org.ID,
		Role:           schema.Role,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create organizationMember")
	}
	return transformersv1.ToOrganizationMemberSchema(ctx, organizationMember)
}

func (c *organizationMemberController) List(ctx *gin.Context, schema *GetOrganizationSchema) ([]*schemasv1.OrganizationMemberSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, org); err != nil {
		return nil, err
	}
	members, err := services.OrganizationMemberService.List(ctx, services.ListOrganizationMemberOption{
		OrganizationId: utils.UintPtr(org.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list organization members")
	}
	return transformersv1.ToOrganizationMemberSchemas(ctx, members)
}

type DeleteOrganizationMemberSchema struct {
	schemasv1.DeleteOrganizationMemberSchema
	GetOrganizationSchema
}

func (c *organizationMemberController) Delete(ctx *gin.Context, schema *DeleteOrganizationMemberSchema) (*schemasv1.OrganizationMemberSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current user")
	}
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	if err = c.canOperate(ctx, org); err != nil {
		return nil, err
	}
	member, err := services.OrganizationMemberService.GetBy(ctx, schema.UserId, org.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get member")
	}
	organizationMember, err := services.OrganizationMemberService.Delete(ctx, member, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create organizationMember")
	}
	return transformersv1.ToOrganizationMemberSchema(ctx, organizationMember)
}
