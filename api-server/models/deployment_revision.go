package models

import "github.com/bentoml/yatai-schemas/modelschemas"

type DeploymentRevision struct {
	BaseModel
	CreatorAssociate
	DeploymentAssociate

	Status modelschemas.DeploymentRevisionStatus `json:"status"`
}

func (s *DeploymentRevision) GetName() string {
	return s.Uid
}

func (s *DeploymentRevision) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeDeploymentRevision
}
