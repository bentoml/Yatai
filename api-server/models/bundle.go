package models

type Bundle struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate
	Description string `json:"description"`
}

func (b *Bundle) GetResourceType() ResourceType {
	return ResourceTypeBundle
}
