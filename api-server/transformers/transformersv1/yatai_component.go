package transformersv1

import (
	"context"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
)

func ToYataiComponentSchema(ctx context.Context, comp *models.YataiComponent) (s *schemasv1.YataiComponentSchema, err error) {
	ss, err := ToYataiComponentSchemas(ctx, []*models.YataiComponent{comp})
	if err != nil {
		return
	}
	return ss[0], nil
}

func ToYataiComponentSchemas(ctx context.Context, comps []*models.YataiComponent) (ss []*schemasv1.YataiComponentSchema, err error) {
	for _, c := range comps {
		ss = append(ss, &schemasv1.YataiComponentSchema{
			Type:    c.Type,
			Release: c.Release,
		})
	}
	return
}
