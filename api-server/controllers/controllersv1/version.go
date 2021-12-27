package controllersv1

import (
	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai/api-server/version"
)

type versionController struct {
	baseController
}

var VersionController = versionController{}

type VersionSchema struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
}

func (c *versionController) GetVersion(ctx *gin.Context) (*VersionSchema, error) {
	return &VersionSchema{
		Version:   version.Version,
		GitCommit: version.GitCommit,
		BuildDate: version.BuildDate,
	}, nil
}
