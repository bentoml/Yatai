package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

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
		orgSchema, err := GetAssociatedNullableOrganizationSchema(ctx, r)
		if err != nil {
			return nil, err
		}

		ss = append(ss, &schemasv1.LabelSchema{
			ResourceSchema: ToResourceSchema(r),
			Organization:   orgSchema,
			Creator:        creatorSchema,
			ResourceType:   r.ResourceType,
			ResourceId:     r.ResourceId,
			Key:            r.Key,
			Value:          r.Value,
		})
	}

	return ss, nil
}
