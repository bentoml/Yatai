package controllersv1

import (
	"net/http"
	"strings"

	"github.com/bentoml/yatai/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type authController struct {
	baseController
}

var AuthController = authController{}

func (*authController) Register(ctx *gin.Context, schema *schemasv1.RegisterUserSchema) (*schemasv1.UserSchema, error) {
	user, err := services.UserService.Create(ctx, services.CreateUserOption{
		Name:      schema.Name,
		FirstName: schema.FirstName,
		LastName:  schema.LastName,
		Email:     utils.StringPtrWithoutEmpty(schema.Email),
		Password:  schema.Password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	err = scookie.SetUsernameToCookie(ctx, user.Name)
	if err != nil {
		return nil, errors.Wrap(err, "set login cookie")
	}
	return transformersv1.ToUserSchema(ctx, user)
}

func (*authController) Login(ctx *gin.Context, schema *schemasv1.LoginUserSchema) (*schemasv1.UserSchema, error) {
	isEmail := strings.Contains(schema.NameOrEmail, "@")
	var err error
	var user *models.User
	if isEmail {
		user, err = services.UserService.GetByEmail(ctx, schema.NameOrEmail)
	} else {
		user, err = services.UserService.GetByName(ctx, schema.NameOrEmail)
	}
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}
	if err = services.UserService.CheckPassword(ctx, user, schema.Password); err != nil {
		return nil, err
	}
	err = scookie.SetUsernameToCookie(ctx, user.Name)
	if err != nil {
		return nil, errors.Wrap(err, "set login cookie")
	}
	redirectUri := ctx.Query("redirect")
	if redirectUri == "" {
		redirectUri = "/"
	}
	ctx.Redirect(http.StatusSeeOther, redirectUri)
	return transformersv1.ToUserSchema(ctx, user)
}

func (*authController) GetCurrentUser(ctx *gin.Context) (*schemasv1.UserSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return transformersv1.ToUserSchema(ctx, user)
}

func (*authController) GenerateApiToken(ctx *gin.Context) (*schemasv1.UserSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	user, err = services.UserService.GenerateApiToken(ctx, user)
	if err != nil {
		err = errors.Wrap(err, "generate api_token failed")
		return nil, err
	}
	return transformersv1.ToUserSchema(ctx, user)
}

func (*authController) DeleteApiToken(ctx *gin.Context) (*schemasv1.UserSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	user, err = services.UserService.DeleteApiToken(ctx, user)
	if err != nil {
		err = errors.Wrap(err, "generate api_token failed")
		return nil, err
	}
	return transformersv1.ToUserSchema(ctx, user)
}
