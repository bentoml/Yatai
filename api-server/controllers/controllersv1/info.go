package controllersv1

import (
	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai/api-server/config"
)

type infoController struct {
	baseController
}

var InfoController = infoController{}

type InfoSchema struct {
	IsSass           bool   `json:"is_sass"`
	SassDomainSuffix string `json:"sass_domain_suffix"`
}

func (c *infoController) GetInfo(ctx *gin.Context) (*InfoSchema, error) {
	return &InfoSchema{
		IsSass:           config.YataiConfig.IsSass,
		SassDomainSuffix: config.YataiConfig.SassDomainSuffix,
	}, nil
}
