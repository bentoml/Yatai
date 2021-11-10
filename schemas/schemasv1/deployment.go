package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentSchema struct {
	ResourceSchema
	Creator *UserSchema                   `json:"creator"`
	Cluster *ClusterFullSchema            `json:"cluster"`
	Status  modelschemas.DeploymentStatus `json:"status"`
	URLs    []string                      `json:"urls"`
}

type DeploymentFullSchema struct {
	DeploymentSchema
	LatestRevision **DeploymentRevisionSchema `json:"latest_revision"`
}

type DeploymentListSchema struct {
	BaseListSchema
	Items []*DeploymentSchema `json:"items"`
}

type UpdateDeploymentSchema struct {
	Targets []*CreateDeploymentTargetSchema `json:"targets"`
}

type CreateDeploymentSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdateDeploymentSchema
}
