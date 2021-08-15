package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type ClusterMemberSchema struct {
	BaseSchema
	Role    modelschemas.MemberRole `json:"role"`
	Creator *UserSchema             `json:"creator"`
	User    UserSchema              `json:"user"`
	Cluster ClusterSchema           `json:"cluster"`
}
