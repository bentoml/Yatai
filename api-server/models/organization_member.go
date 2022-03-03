package models

import "github.com/bentoml/yatai-schemas/modelschemas"

type organizationMemberModel struct{}

var OrganizationMemberModel = organizationMemberModel{}

type OrganizationMember struct {
	BaseModel
	CreatorAssociate
	UserAssociate
	OrganizationAssociate

	Role modelschemas.MemberRole `json:"role"`
}
