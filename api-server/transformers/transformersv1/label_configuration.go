package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToLabelConfigurationSchema(ctx context.Context, labelConfig *models.LabelConfiguration) (*schemasv1.LabelConfigurationSchema, error) {
	schemas, err := ToLabelConfigurationSchemas(ctx, []*models.LabelConfiguration{labelConfig})
	if err != nil {
		return nil, err
	}
	return schemas[0], nil
}

func ToLabelConfigurationSchemas(ctx context.Context, labelConfigurations []*models.LabelConfiguration) ([]*schemasv1.LabelConfigurationSchema, error) {
	schemas := make([]*schemasv1.LabelConfigurationSchema, 0, len(labelConfigurations))
	for _, labelConfiguration := range labelConfigurations {
		schemas = append(schemas, &schemasv1.LabelConfigurationSchema{
			BaseSchema: ToBaseSchema(labelConfiguration),
			Key:        labelConfiguration.Key,
			Info:       labelConfiguration.Info,
		})
	}
	return schemas, nil
}
