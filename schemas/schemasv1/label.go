package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelSchema struct{
	ResourceSchema
	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceUid string
	Key			string `json:"key"`
	Value 		string `json:"value`
}

type LabelListSchema struct {
	BaseListSchema
	Items []*LabelSchema `json:"labels"`
}