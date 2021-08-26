package modelschemas

type ResourceType string

const (
	ResourceTypeUser         ResourceType = "user"
	ResourceTypeOrganization ResourceType = "organization"
	ResourceTypeCluster      ResourceType = "cluster"
	ResourceTypeBento        ResourceType = "bento"
	ResourceTypeBentoVersion ResourceType = "bento_version"
	ResourceTypeDeployment   ResourceType = "deployment"
)
