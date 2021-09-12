package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentSnapshotSchema struct {
	ResourceSchema
	Creator      *UserSchema                                 `json:"creator"`
	Type         modelschemas.DeploymentSnapshotType         `json:"type"`
	Status       modelschemas.DeploymentSnapshotStatus       `json:"status"`
	BentoVersion *BentoVersionFullSchema                     `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentSnapshotCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentSnapshotConfig      `json:"config"`
}

type DeploymentSnapshotListSchema struct {
	BaseListSchema
	Items []*DeploymentSnapshotSchema `json:"items"`
}
