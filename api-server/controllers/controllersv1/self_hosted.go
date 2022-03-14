package controllersv1

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/common/utils"
)

type selfHostedController struct {
	baseController
}

type SetupSchema struct {
	schemasv1.RegisterUserSchema
	Token string `json:"token"`
}

var SelfHostedController = selfHostedController{}

func (*selfHostedController) Setup(ctx *gin.Context, schema *SetupSchema) (*schemasv1.UserSchema, error) {
	/*
	* Setup default admin, org, cluster route for self hosted yatai
	*
	* This route will setup default admin, org, and cluster only:
	* 1. is NOT in sass mode
	* 2. the token in request is the same as the token in config
	* 3. There is no user in db
	* 4. There is no org in db
	* 5. There is no cluster in db
	* If any of the above condition is not met, this route will return error
	*
	* this endpoint will:
	* 1. create a user with admin permission,
	* 2. create a org,and add the user to the org,
	* 3. create a cluster, and add the org to the cluster,
	 */

	if config.YataiConfig.IsSass {
		return nil, errors.New("admin user registration is not allowed in sass mode")
	}
	if config.YataiConfig.InitializationToken == "" {
		return nil, errors.New("initialization token is not set")
	}
	if schema.Token != config.YataiConfig.InitializationToken {
		return nil, errors.New("invalid token")
	}

	users, _, err := services.UserService.List(ctx, services.ListUserOption{
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}

	for _, user_ := range users {
		if user_.Email != nil && *user_.Email != "" {
			return nil, errors.New("the setup has been completed long ago")
		}
	}

	if len(users) == 0 {
		return nil, errors.New("no user found, seems not complete self hosted setup")
	}

	user := users[0]

	user, err = services.UserService.ForceUpdatePassword(ctx, user, schema.Password)
	if err != nil {
		return nil, errors.Wrap(err, "update admin user password")
	}
	email := &schema.Email
	user, err = services.UserService.Update(ctx, user, services.UpdateUserOption{
		Email:     &email,
		Name:      &schema.Name,
		FirstName: &schema.FirstName,
		LastName:  &schema.LastName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update admin user")
	}

	err = scookie.SetUsernameToCookie(ctx, user.Name)
	if err != nil {
		return nil, errors.Wrap(err, "set login cookie")
	}
	return transformersv1.ToUserSchema(ctx, user)
}
