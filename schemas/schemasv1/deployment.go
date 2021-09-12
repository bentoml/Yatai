package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentSchema struct {
	ResourceSchema
	Creator *UserSchema                   `json:"creator"`
	Cluster *ClusterFullSchema            `json:"cluster"`
	Status  modelschemas.DeploymentStatus `json:"status"`
	URLs    []string                      `json:"urls"`
}

type DeploymentListSchema struct {
	BaseListSchema
	Items []*DeploymentSchema `json:"items"`
}

type UpdateDeploymentSchema struct {
	Type         modelschemas.DeploymentSnapshotType         `json:"type"`
	BentoName    string                                      `json:"bento_name"`
	BentoVersion string                                      `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentSnapshotCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentSnapshotConfig      `json:"config"`
}

type CreateDeploymentSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdateDeploymentSchema
}
