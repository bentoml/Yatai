package modelschemas

type ResourceType string

const (
	ResourceTypeUser               ResourceType = "user"
	ResourceTypeOrganization       ResourceType = "organization"
	ResourceTypeCluster            ResourceType = "cluster"
	ResourceTypeBento              ResourceType = "bento"
	ResourceTypeBentoVersion       ResourceType = "bento_version"
	ResourceTypeDeployment         ResourceType = "deployment"
	ResourceTypeDeploymentRevision ResourceType = "deployment_revision"
	ResourceTypeTerminalRecord     ResourceType = "terminal_record"
	ResourceTypeModel              ResourceType = "model"
	ResourceTypeModelVersion       ResourceType = "model_version"
	ResourceTypeLabel              ResourceType = "label"
	ResourceTypeApiToken           ResourceType = "api_token"
)
