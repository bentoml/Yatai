package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type Bundle struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *Bundle) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeBundle
}
