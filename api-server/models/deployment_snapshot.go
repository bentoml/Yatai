package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type DeploymentSnapshot struct {
	BaseModel
	CreatorAssociate
	DeploymentAssociate
	BentoVersionAssociate

	Type        modelschemas.DeploymentSnapshotType         `json:"type"`
	Status      modelschemas.DeploymentSnapshotStatus       `json:"status"`
	CanaryRules *modelschemas.DeploymentSnapshotCanaryRules `json:"canary_rules"`
	Config      *modelschemas.DeploymentSnapshotConfig      `json:"config"`
}

func (s *DeploymentSnapshot) GetName() string {
	return s.Uid
}

func (s *DeploymentSnapshot) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeDeploymentSnapshot
}
