package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelSchema struct {
	ResourceSchema
	Organization *OrganizationSchema       `json:"organization"`
	Creator      *UserSchema               `json:"creator"`
	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceId   uint                      `json:"resource_id"`
	Key          string                    `json:"key"`
	Value        string                    `json:"value"`
}

type LabelListSchema struct {
	BaseListSchema
	Items []*LabelSchema `json:"labels"`
}

type CreateLabelSchema struct {
	LabelKey   string `json:"key"`
	LabelValue string `json:"value"`
}

type UpdateLabelSchema struct {
	LabelValue string `json:"value"`
}
