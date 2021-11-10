package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

// TODO

func ToLabelSchema(ctx context.Context, label *models.Label) (*schemasv1.LabelSchema, error) {
	ss, err := ToLabelSchemas(ctx, []*models.Label{label})
	if err != nil {
		return nil, err
	}
	return ss[0], nil
}

func ToLabelSchemas(ctx context.Context, labels []*models.Label) ([]*schemasv1.LabelSchema, error) {
	ss := make([]*schemasv1.LabelSchema, 0, len(labels))
	for _, r := range labels {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, r)
		if err != nil {
			return nil, err
		}
		resource, err := services.ResourceService.Get(ctx, r.ResourceType, r.ResourceId)
		if err != nil && !utils.IsNotFound(err) {
			return nil, err
		}
		var rs *schemasv1.ResourceSchema
		if !utils.IsNotFound(err) {
			rs_ := ToResourceSchema(resource)
			rs = &rs_
		}
		ss = append(ss, &schemasv1.LabelSchema{
			ResourceSchema: ToResourceSchema(r),
			Creator: 		creatorSchema,
			Resource: 		rs,
			Key:			r.Key,
			Value:			r.Value,
		})
	}

	return ss, nil
}