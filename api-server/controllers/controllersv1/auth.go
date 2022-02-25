package controllersv1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/scookie"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"
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
		return nil, errors.New("invalid username or password")
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

func (*authController) ResetPassword(ctx *gin.Context, schema *schemasv1.ResetPasswordSchema) (*schemasv1.UserSchema, error) {
	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	user, err := services.UserService.UpdatePassword(ctx, currentUser, schema.CurrentPassword, schema.NewPassword)
	if err != nil {
		return nil, err
	}

	return transformersv1.ToUserSchema(ctx, user)
}

func (*authController) RegisterAdminUser(ctx *gin.Context, schema *schemasv1.RegisterUserSchema) (*schemasv1.UserSchema, error) {
	if config.YataiConfig.IsSass {
		return nil, errors.New("admin user registration is not allowed in sass mode")
	}

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
	orgs, total, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list organizations")
	}
	if total == 0 {
		return nil, errors.New("no organization found")
	}
	defaultOrg := orgs[0]

	_, err = services.OrganizationMemberService.Create(ctx, user.ID, services.CreateOrganizationMemberOption{
		Role:           modelschemas.MemberRoleAdmin,
		OrganizationId: defaultOrg.ID,
		UserId:         user.ID,
		CreatorId:      user.ID,
	})

	clusters, total, err := services.ClusterService.List(ctx, services.ListClusterOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list clusters")
	}
	if total == 0 {
		return nil, errors.New("no cluster found")
	}
	defaultCluster := clusters[0]

	_, err = services.ClusterMemberService.Create(ctx, user.ID, services.CreateClusterMemberOption{
		Role:      modelschemas.MemberRoleAdmin,
		ClusterId: defaultCluster.ID,
		UserId:    user.ID,
		CreatorId: user.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create cluster member")
	}

	err = scookie.SetUsernameToCookie(ctx, user.Name)
	if err != nil {
		return nil, errors.Wrap(err, "set login cookie")
	}
	return transformersv1.ToUserSchema(ctx, user)
}
