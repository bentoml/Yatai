package models

type Bundle struct {
	ResourceMixin
	CreatorAssociate
	ClusterAssociate
	Description string `json:"description"`
}

func (b *Bundle) GetResourceType() ResourceType {
	return ResourceTypeBundle
}
