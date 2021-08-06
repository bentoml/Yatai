package models

type UserGroup struct {
	ResourceMixin
	OrganizationAssociate
	CreatorAssociate
}
