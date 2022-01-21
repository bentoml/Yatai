package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type Deployment struct {
	ResourceMixin
	CreatorAssociate
	ClusterAssociate

	Description     string                        `json:"description"`
	Status          modelschemas.DeploymentStatus `json:"status"`
	StatusSyncingAt *time.Time                    `json:"status_syncing_at"`
	StatusUpdatedAt *time.Time                    `json:"status_updated_at"`
	KubeDeployToken string                        `json:"kube_deploy_token"`
	KubeNamespace   string                        `json:"kube_namespace"`
}

func (d *Deployment) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeDeployment
}
