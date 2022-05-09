package controllersv1

import (
	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/version"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/envars"
)

type versionController struct {
	baseController
}

var VersionController = versionController{}

func (c *versionController) GetVersion(ctx *gin.Context) (*schemasv1.VersionSchema, error) {
	// Set YATAI_VERSION
	envars.SetIfNotExists(consts.EnvYataiVersion, version.Version)

	return &schemasv1.VersionSchema{
		Version:   version.Version,
		GitCommit: version.GitCommit,
		BuildDate: version.BuildDate,
	}, nil
}
