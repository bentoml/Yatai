package models

type Organization struct {
	ResourceMixin
	CreatorAssociate

	Description string `json:"description"`
}

func (o *Organization) GetResourceType() ResourceType {
	return ResourceTypeOrganization
}
