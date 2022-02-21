package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToOrganizationMemberSchema(ctx context.Context, member *models.OrganizationMember) (*schemasv1.OrganizationMemberSchema, error) {
	if member == nil {
		return nil, nil
	}
	ss, err := ToOrganizationMemberSchemas(ctx, []*models.OrganizationMember{member})
	if err != nil {
		return nil, errors.Wrap(err, "ToOrganizationMemberSchemas")
	}
	return ss[0], nil
}

func ToOrganizationMemberSchemas(ctx context.Context, members []*models.OrganizationMember) ([]*schemasv1.OrganizationMemberSchema, error) {
	res := make([]*schemasv1.OrganizationMemberSchema, 0, len(members))
	for _, member := range members {
		creator, err := services.UserService.GetAssociatedCreator(ctx, member)
		if err != nil {
			return nil, errors.Wrap(err, "get organization member associated creator")
		}
		creatorSchema, err := ToUserSchema(ctx, creator)
		if err != nil {
			return nil, errors.Wrap(err, "ToUserSchema")
		}

		user, err := services.UserService.GetAssociatedUser(ctx, member)
		if err != nil {
			return nil, errors.Wrap(err, "get organization member associated user")
		}
		userSchema, err := ToUserSchema(ctx, user)
		if err != nil {
			return nil, errors.Wrap(err, "ToUserSchema")
		}

		org, err := services.OrganizationService.GetAssociatedOrganization(ctx, member)
		if err != nil {
			return nil, errors.Wrap(err, "get organization member associated organization")
		}
		orgSchema, err := ToOrganizationSchema(ctx, org)
		if err != nil {
			return nil, errors.Wrap(err, "ToOrganizationSchema")
		}

		res = append(res, &schemasv1.OrganizationMemberSchema{
			BaseSchema: schemasv1.BaseSchema{
				DeletedAt: &member.DeletedAt.Time,
			},
			Creator:      creatorSchema,
			User:         *userSchema,
			Organization: *orgSchema,
			Role:         member.Role,
		})
	}
	return res, nil
}
