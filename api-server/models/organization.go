package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type Organization struct {
	ResourceMixin
	CreatorAssociate

	Description string                                 `json:"description"`
	Config      *modelschemas.OrganizationConfigSchema `json:"config"`
}

func (o *Organization) GetResourceType() ResourceType {
	return ResourceTypeOrganization
}
