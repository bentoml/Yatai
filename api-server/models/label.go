package models

import "github.com/bentoml/yatai-schemas/modelschemas"

type Label struct {
	BaseModel
	CreatorAssociate
	OrganizationAssociate

	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceId   uint                      `json:"resource_id"`

	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *Label) GetName() string {
	return r.Uid
}

func (r *Label) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeLabel
}
