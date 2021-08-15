package controllersv1

import (
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type clusterMemberController struct {
	clusterController
}

var ClusterMemberController = clusterMemberController{}

type CreateClusterMembersSchema struct {
	schemasv1.CreateMembersSchema
	GetClusterSchema
}

func (c *clusterMemberController) Create(ctx *gin.Context, schema *CreateClusterMembersSchema) ([]*schemasv1.ClusterMemberSchema, error) {
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current user")
	}
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	users, err := services.UserService.ListByNames(ctx, schema.Usernames)
	if err != nil {
		return nil, err
	}
	res := make([]*schemasv1.ClusterMemberSchema, 0, len(users))
	for _, u := range users {
		clusterMember, err := services.ClusterMemberService.Create(ctx, currentUser.ID, services.CreateClusterMemberOption{
			CreatorId: currentUser.ID,
			UserId:    u.ID,
			ClusterId: cluster.ID,
			Role:      schema.Role,
		})
		if err != nil {
			return nil, errors.Wrap(err, "create clusterMember")
		}
		s, err := transformersv1.ToClusterMemberSchema(ctx, clusterMember)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

func (c *clusterMemberController) List(ctx *gin.Context, schema *GetClusterSchema) ([]*schemasv1.ClusterMemberSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, cluster); err != nil {
		return nil, err
	}
	members, err := services.ClusterMemberService.List(ctx, services.ListClusterMemberOption{
		ClusterId: utils.UintPtr(cluster.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list organization members")
	}
	return transformersv1.ToClusterMemberSchemas(ctx, members)
}

type DeleteClusterMemberSchema struct {
	schemasv1.DeleteMemberSchema
	GetClusterSchema
}

func (c *clusterMemberController) Delete(ctx *gin.Context, schema *DeleteClusterMemberSchema) (*schemasv1.ClusterMemberSchema, error) {
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current user")
	}
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	if err = c.canOperate(ctx, cluster); err != nil {
		return nil, err
	}
	user, err := services.UserService.GetByName(ctx, schema.Username)
	if err != nil {
		return nil, err
	}
	member, err := services.ClusterMemberService.GetBy(ctx, user.ID, cluster.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get member")
	}
	clusterMember, err := services.ClusterMemberService.Delete(ctx, member, currentUser.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create clusterMember")
	}
	return transformersv1.ToClusterMemberSchema(ctx, clusterMember)
}
