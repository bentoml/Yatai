package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type Bento struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *Bento) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeBento
}
