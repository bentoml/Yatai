package schemasv1

import (
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type ClusterSchema struct {
	ResourceSchema
	Creator     *UserSchema `json:"creator"`
	Description string      `json:"description"`
}

type ClusterListSchema struct {
	BaseListSchema
	Items []*ClusterSchema `json:"items"`
}

type ClusterFullSchema struct {
	ClusterSchema
	Organization *OrganizationSchema                `json:"organization"`
	KubeConfig   *string                            `json:"kube_config"`
	Config       **modelschemas.ClusterConfigSchema `json:"config"`
}

type UpdateClusterSchema struct {
	Description *string                            `json:"description"`
	KubeConfig  *string                            `json:"kube_config"`
	Config      **modelschemas.ClusterConfigSchema `json:"config"`
}

type CreateClusterSchema struct {
	Description string                            `json:"description"`
	KubeConfig  string                            `json:"kube_config"`
	Config      *modelschemas.ClusterConfigSchema `json:"config"`
	Name        string                            `json:"name"`
}
