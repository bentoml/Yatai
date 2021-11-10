package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentTargetSchema struct {
	ResourceSchema
	Creator      *UserSchema                               `json:"creator"`
	Type         modelschemas.DeploymentTargetType         `json:"type"`
	BentoVersion *BentoVersionFullSchema                   `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentTargetCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentTargetConfig      `json:"config"`
}

type DeploymentTargetListSchema struct {
	BaseListSchema
	Items []*DeploymentTargetSchema `json:"items"`
}

type CreateDeploymentTargetSchema struct {
	Type         modelschemas.DeploymentTargetType         `json:"type"`
	BentoName    string                                    `json:"bento_name"`
	BentoVersion string                                    `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentTargetCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentTargetConfig      `json:"config"`
}
