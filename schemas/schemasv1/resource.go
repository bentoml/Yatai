package schemasv1

type ResourceType string

const (
	ResourceTypeUser          ResourceType = "user"
	ResourceTypeOrganization  ResourceType = "organization"
	ResourceTypeCluster       ResourceType = "cluster"
	ResourceTypeBundle        ResourceType = "bundle"
	ResourceTypeBundleVersion ResourceType = "bundle_version"
	ResourceTypeDeployment    ResourceType = "deployment"
)

type ResourceSchema struct {
	BaseSchema
	Name         string       `json:"name"`
	ResourceType ResourceType `json:"resource_type" enum:"user,organization,cluster,bundle,bundle_version,deployment"`
}

func (s *ResourceSchema) TypeName() string {
	return string(s.ResourceType)
}
