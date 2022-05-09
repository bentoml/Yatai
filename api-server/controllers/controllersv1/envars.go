package controllersv1

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
)

type envarsController struct {
	baseController
}

type EnvarsSchema struct{}

var EnvarsController = envarsController{}

func (c *envarsController) SetEnvars(ctx *gin.Context) (*EnvarsSchema, error) {

	// get versions from versionControllers
	versionSchema, err := VersionController.GetVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get version")
	}
	os.Setenv(consts.EnvYataiVersion, versionSchema.Version)

	// get organization uid from current user
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organizationModel, err := services.OrganizationService.GetUserOrganization(ctx, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get user org")
	}
	os.Setenv(consts.EnvYataiOrgUID, organizationModel.GetUid())

	// get major cluster uid from current user
	clusterModel, err := services.OrganizationService.GetMajorCluster(ctx, organizationModel)
	if err != nil {
		return nil, errors.Wrap(err, "get major cluster")
	}
	os.Setenv(consts.EnvYataiClusterUID, clusterModel.GetUid())

	// get deployments uid from given cluster

	noop := &EnvarsSchema{}

	return noop, nil
}
