package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type OrganizationMemberSchema struct {
	BaseSchema
	Role         modelschemas.MemberRole `json:"role"`
	Creator      *UserSchema             `json:"creator"`
	User         UserSchema              `json:"user"`
	Organization OrganizationSchema      `json:"organization"`
}

type CreateOrganizationMemberSchema struct {
	UserId uint                    `json:"user_id"`
	Role   modelschemas.MemberRole `json:"role" enum:"guest,developer,admin"`
}

type DeleteOrganizationMemberSchema struct {
	UserId uint `json:"user_id"`
}
