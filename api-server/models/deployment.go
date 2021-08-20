package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type DeploymentStatus string

const (
	DeploymentStatusUnknown     DeploymentStatus = "unknown"
	DeploymentStatusNonDeployed DeploymentStatus = "non-deployed"
	DeploymentStatusRunning     DeploymentStatus = "running"
	DeploymentStatusUnhealthy   DeploymentStatus = "unhealthy"
	DeploymentStatusFailed      DeploymentStatus = "failed"
	DeploymentStatusDeploying   DeploymentStatus = "deploying"
)

type Deployment struct {
	ResourceMixin
	CreatorAssociate
	ClusterAssociate

	Description     string           `json:"description"`
	Status          DeploymentStatus `json:"status"`
	StatusSyncingAt *time.Time       `json:"status_syncing_at"`
	StatusUpdatedAt *time.Time       `json:"status_updated_at"`
}

func (d *Deployment) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeDeployment
}
