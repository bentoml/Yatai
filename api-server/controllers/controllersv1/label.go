package controllersv1 

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type labelController struct {
	baseController
}

var labelController = labelController{}

type GetLabelSchema struct {
	Uid string `path:"uid"`
}

func (s *GetLabelSchema) GetLabel(ctx context.Context) (*models.Label, error) {
	return services.LabelService.GetByUid(ctx, s.Uid)
}

//TODO