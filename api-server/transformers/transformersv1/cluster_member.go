package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/pkg/errors"
)

func ToClusterMemberSchema(ctx context.Context, member *models.ClusterMember) (*schemasv1.ClusterMemberSchema, error) {
	if member == nil {
		return nil, nil
	}
	ss, err := ToClusterMemberSchemas(ctx, []*models.ClusterMember{member})
	if err != nil {
		return nil, errors.Wrap(err, "ToClusterMemberSchemas")
	}
	return ss[0], nil
}

func ToClusterMemberSchemas(ctx context.Context, members []*models.ClusterMember) ([]*schemasv1.ClusterMemberSchema, error) {
	res := make([]*schemasv1.ClusterMemberSchema, 0, len(members))
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

		cluster, err := services.ClusterService.GetAssociatedCluster(ctx, member)
		if err != nil {
			return nil, errors.Wrap(err, "get organization member associated organization")
		}
		clusterSchema, err := ToClusterSchema(ctx, cluster)
		if err != nil {
			return nil, errors.Wrap(err, "ToClusterSchema")
		}

		res = append(res, &schemasv1.ClusterMemberSchema{
			Creator: creatorSchema,
			User:    *userSchema,
			Cluster: *clusterSchema,
			Role:    member.Role,
		})
	}
	return res, nil
}
