package controllersv1

import (
	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/version"
)

type versionController struct {
	// nolint: unused
	baseController
}

var VersionController = versionController{}

func (c *versionController) GetVersion(ctx *gin.Context) (*schemasv1.VersionSchema, error) {
	return &schemasv1.VersionSchema{
		Version:   version.Version,
		GitCommit: version.GitCommit,
		BuildDate: version.BuildDate,
	}, nil
}
