package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentTarget struct {
	BaseModel
	CreatorAssociate
	DeploymentAssociate
	DeploymentRevisionAssociate
	BentoAssociate

	Type        modelschemas.DeploymentTargetType         `json:"type"`
	CanaryRules *modelschemas.DeploymentTargetCanaryRules `json:"canary_rules"`
	Config      *modelschemas.DeploymentTargetConfig      `json:"config"`
}

func (s *DeploymentTarget) GetName() string {
	return s.Uid
}

func (s *DeploymentTarget) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeDeploymentRevision
}
