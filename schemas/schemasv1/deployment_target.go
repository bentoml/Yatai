package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentTargetTypeSchema struct {
	Type modelschemas.DeploymentTargetType `json:"type" enum:"stable,canary"`
}

type DeploymentTargetSchema struct {
	ResourceSchema
	DeploymentTargetTypeSchema
	Creator      *UserSchema                               `json:"creator"`
	BentoVersion *BentoVersionFullSchema                   `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentTargetCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentTargetConfig      `json:"config"`
}

type DeploymentTargetListSchema struct {
	BaseListSchema
	Items []*DeploymentTargetSchema `json:"items"`
}

type  struct {
	DeploymentTargetTypeSchema
	BentoName    string                                    `json:"bento_name"`
	BentoVersion string                                    `json:"bento_version"`
	CanaryRules  *modelschemas.DeploymentTargetCanaryRules `json:"canary_rules"`
	Config       *modelschemas.DeploymentTargetConfig      `json:"config"`
}
