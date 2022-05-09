package controllersv1

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/common/consts"
)

type envarsController struct {
	baseController
}

type NoopSchema struct{}

var EnvarsController = envarsController{}

func (c *envarsController) SetEnvars(ctx *gin.Context, schema *GetDeploymentSchema) (*NoopSchema, error) {
	// get versions from versionControllers
	versionSchema, err := VersionController.GetVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get version")
	}
	os.Setenv(consts.EnvYataiVersion, versionSchema.Version)

	// get organization uid
	organizationModel, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get user org uid")
	}
	os.Setenv(consts.EnvYataiOrgUID, organizationModel.GetUid())

	// get cluster uid from current user
	clusterModel, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster uid")
	}
	os.Setenv(consts.EnvYataiClusterUID, clusterModel.GetUid())

	// get deployments uid from given cluster
	deploymentModel, err := schema.GetDeployment(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get deployment uid")
	}
	os.Setenv(consts.EnvYataiDeploymentUID, deploymentModel.GetUid())

	return &NoopSchema{}, nil
}
