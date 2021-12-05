package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type BentoRepository struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *BentoRepository) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeBentoRepository
}
