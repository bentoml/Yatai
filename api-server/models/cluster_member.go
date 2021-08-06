package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type ClusterMember struct {
	BaseModel
	CreatorAssociate
	UserAssociate
	ClusterAssociate

	Role modelschemas.MemberRole `json:"role"`
}
