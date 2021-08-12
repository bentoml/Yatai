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

type CreateOrganizationMembersSchema struct {
	schemasv1.CreateMembersSchema
	GetOrganizationSchema
}

func (c *organizationMemberController) Create(ctx *gin.Context, schema *CreateOrganizationMembersSchema) ([]*schemasv1.OrganizationMemberSchema, error) {
	currentUser, err := services.GetCurrentUser(ctx)
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
	users, err := services.UserService.ListByNames(ctx, schema.Usernames)
	if err != nil {
		return nil, err
	}
	res := make([]*schemasv1.OrganizationMemberSchema, 0, len(users))
	for _, u := range users {
		organizationMember, err := services.OrganizationMemberService.Create(ctx, currentUser.ID, services.CreateOrganizationMemberOption{
			CreatorId:      currentUser.ID,
			UserId:         u.ID,
			OrganizationId: org.ID,
			Role:           schema.Role,
		})
		if err != nil {
			return nil, errors.Wrap(err, "create organizationMember")
		}
		s, err := transformersv1.ToOrganizationMemberSchema(ctx, organizationMember)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
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
	schemasv1.DeleteMemberSchema
	GetOrganizationSchema
}

func (c *organizationMemberController) Delete(ctx *gin.Context, schema *DeleteOrganizationMemberSchema) (*schemasv1.OrganizationMemberSchema, error) {
	currentUser, err := services.GetCurrentUser(ctx)
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
	user, err := services.UserService.GetByName(ctx, schema.Username)
	if err != nil {
		return nil, err
	}
	member, err := services.OrganizationMemberService.GetBy(ctx, user.ID, org.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get member")
	}
	organizationMember, err := services.OrganizationMemberService.Delete(ctx, member, currentUser.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create organizationMember")
	}
	return transformersv1.ToOrganizationMemberSchema(ctx, organizationMember)
}
