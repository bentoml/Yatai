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
	IsSaas           bool   `json:"is_saas"`
	SaasDomainSuffix string `json:"saas_domain_suffix"`
}

func (c *infoController) GetInfo(ctx *gin.Context) (*InfoSchema, error) {
	return &InfoSchema{
		IsSaas:           config.YataiConfig.IsSaaS,
		SaasDomainSuffix: config.YataiConfig.SaasDomainSuffix,
	}, nil
}
