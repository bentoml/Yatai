package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type Cluster struct {
	ResourceMixin
	CreatorAssociate
	OrganizationAssociate

	Description string                            `json:"description"`
	KubeConfig  string                            `json:"kube_config"`
	Config      *modelschemas.ClusterConfigSchema `json:"config"`
}

func (c *Cluster) GetResourceType() ResourceType {
	return ResourceTypeCluster
}
