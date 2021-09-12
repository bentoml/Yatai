package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type Model struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *Model) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeModel
}
