package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type ClusterMemberSchema struct {
	BaseSchema
	Role    modelschemas.MemberRole `json:"role"`
	Creator *UserSchema             `json:"creator"`
	User    UserSchema              `json:"user"`
	Cluster ClusterSchema           `json:"cluster"`
}

type CreateClusterMemberSchema struct {
	UserId    uint                    `json:"user_id"`
	ClusterId uint                    `json:"cluster_id"`
	Role      modelschemas.MemberRole `json:"role" enum:"guest,developer,admin"`
}

type DeleteClusterMemberSchema struct {
	UserId    uint `json:"user_id"`
	ClusterId uint `json:"cluster_id"`
}
