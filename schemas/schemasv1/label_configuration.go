package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelConfigurationSchema struct {
	BaseSchema

	Key  string                               `json:"key"`
	Info *modelschemas.LabelConfigurationInfo `json:"info"`
}

type LabelConfigurationListSchema struct {
	BaseListSchema

	Items []*LabelConfigurationSchema `json:"items"`
}

type CreateLabelConfigurationSchema struct {
	Key  string                               `json:"key"`
	Info *modelschemas.LabelConfigurationInfo `json:"info"`
}

type UpdateLabelConfigurationSchema struct {
	Info *modelschemas.LabelConfigurationInfo `json:"info"`
}
