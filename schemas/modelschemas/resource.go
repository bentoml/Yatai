package modelschemas

type ResourceType string

const (
	ResourceTypeUser          ResourceType = "user"
	ResourceTypeOrganization  ResourceType = "organization"
	ResourceTypeCluster       ResourceType = "cluster"
	ResourceTypeBundle        ResourceType = "bundle"
	ResourceTypeBundleVersion ResourceType = "bundle_version"
	ResourceTypeDeployment    ResourceType = "deployment"
)
