package controllersv1

import (
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
	if schema.Token != config.YataiConfig.InitializationToken {
		return nil, errors.New("invalid token")
	}
	var adminUser *models.User
	users, total, err := services.UserService.List(ctx, services.ListUserOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}
	if total == 0 {
		// Create default admin user
		adminUser, err = services.UserService.Create(ctx, services.CreateUserOption{
			Name:      schema.Name,
			FirstName: schema.FirstName,
			LastName:  schema.LastName,
			Email:     utils.StringPtrWithoutEmpty(schema.Email),
			Password:  schema.Password,
		})
		if err != nil {
			return nil, errors.Wrap(err, "create user")
		}
	} else {
		userInDB := users[0]
		if userInDB.Name == schema.Name {
			adminUser = userInDB
		} else {
			return nil, errors.New("admin user already exists")
		}
	}

	var defaultOrg *models.Organization
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
		// Create default org
		defaultOrg, err = services.OrganizationService.Create(ctx, services.CreateOrganizationOption{
			CreatorId: adminUser.ID,
			Name:      "default",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default organization")
		}
		_, err = services.OrganizationMemberService.Create(ctx, adminUser.ID, services.CreateOrganizationMemberOption{
			CreatorId:      adminUser.ID,
			UserId:         adminUser.ID,
			OrganizationId: defaultOrg.ID,
			Role:           modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default organization member")
		}
	} else {
		defaultOrg = orgs[0]
	}

	var defaultCluster *models.Cluster
	_, total, err = services.ClusterService.List(ctx, services.ListClusterOption{
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
		// create default cluster
		defaultCluster, err = services.ClusterService.Create(ctx, services.CreateClusterOption{
			CreatorId:      adminUser.ID,
			OrganizationId: defaultOrg.ID,
			Name:           "default",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default cluster")
		}
		_, err = services.ClusterMemberService.Create(ctx, adminUser.ID, services.CreateClusterMemberOption{
			CreatorId: adminUser.ID,
			UserId:    adminUser.ID,
			ClusterId: defaultCluster.ID,
			Role:      modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "create default cluster member")
		}
	}

	err = scookie.SetUsernameToCookie(ctx, adminUser.Name)
	if err != nil {
		return nil, errors.Wrap(err, "set login cookie")
	}
	return transformersv1.ToUserSchema(ctx, adminUser)
}
