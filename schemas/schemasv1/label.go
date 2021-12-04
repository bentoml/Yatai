package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelSchema struct {
	ResourceSchema
	Organization *OrganizationSchema       `json:"organization"`
	Creator      *UserSchema               `json:"creator"`
	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceUid  string                    `json:"resource_uid"`
	Key          string                    `json:"key"`
	Value        string                    `json:"value"`
}

type LabelListSchema struct {
	BaseListSchema
	Items []*LabelSchema `json:"labels"`
}

type LabelWithValuesSchema struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type CreateLabelSchema struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdateLabelSchema struct {
	Value string `json:"value"`
}
