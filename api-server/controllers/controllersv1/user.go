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

type userController struct {
	baseController
}

var UserController = userController{}

type GetUserSchema struct {
	UserName string `path:"userName"`
}

func (s *GetUserSchema) GetUser(ctx context.Context) (*models.User, error) {
	user, err := services.UserService.GetByName(ctx, s.UserName)
	if err != nil {
		return nil, errors.Wrapf(err, "get user %s", s.UserName)
	}
	return user, nil
}

func (c *userController) Get(ctx *gin.Context, schema *GetUserSchema) (*schemasv1.UserSchema, error) {
	user, err := schema.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToUserSchema(ctx, user)
}

func (c *userController) List(ctx *gin.Context, schema *schemasv1.ListQuerySchema) (*schemasv1.UserListSchema, error) {
	users, total, err := services.UserService.List(ctx, services.ListUserOption{BaseListOption: services.BaseListOption{
		Start:  utils.UintPtr(schema.Start),
		Count:  utils.UintPtr(schema.Count),
		Search: schema.Search,
	}})
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}
	userSchemas, err := transformersv1.ToUserSchemas(ctx, users)
	return &schemasv1.UserListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: userSchemas,
	}, err
}
