package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type ModelRepository struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *ModelRepository) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeModelRepository
}
