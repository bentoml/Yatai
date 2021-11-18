package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelSchema struct {
	ResourceSchema
	Creator      *UserSchema `json:"creator"`
	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceId  uint `json:"resource_id"`
	Key          string `json:"key"`
	Value        string `json:"value`
}

type LabelListSchema struct {
	BaseListSchema
	Items []*LabelSchema `json:"labels"`
}
