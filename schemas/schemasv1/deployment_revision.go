package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentRevisionSchema struct {
	ResourceSchema
	Creator *UserSchema                           `json:"creator"`
	Status  modelschemas.DeploymentRevisionStatus `json:"status" enum:"active,inactive"`
	Targets []*DeploymentTargetSchema             `json:"targets"`
}

type DeploymentRevisionListSchema struct {
	BaseListSchema
	Items []*DeploymentRevisionSchema `json:"items"`
}
