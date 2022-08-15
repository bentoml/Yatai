package models

import (
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type YataiComponent struct {
	ResourceMixin
	CreatorAssociate
	ClusterAssociate
	OrganizationAssociate

	Version           string                                     `json:"version"`
	KubeNamespace     string                                     `json:"kube_namespace"`
	Description       string                                     `json:"description"`
	Manifest          *modelschemas.YataiComponentManifestSchema `json:"manifest" type:"jsonb"`
	LatestInstalledAt *time.Time                                 `json:"latest_installed_at"`
	LatestHeartbeatAt *time.Time                                 `json:"latest_heartbeat_at"`
}

func (d *YataiComponent) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeYataiComponent
}
