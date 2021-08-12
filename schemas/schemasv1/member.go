package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type CreateMembersSchema struct {
	Usernames []string                `json:"usernames"`
	Role      modelschemas.MemberRole `json:"role" enum:"guest,developer,admin"`
}

type DeleteMemberSchema struct {
	Username string `json:"username"`
}
