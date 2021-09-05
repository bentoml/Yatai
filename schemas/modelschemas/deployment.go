package modelschemas

type DeploymentStatus string

const (
	DeploymentStatusUnknown     DeploymentStatus = "unknown"
	DeploymentStatusNonDeployed DeploymentStatus = "non-deployed"
	DeploymentStatusRunning     DeploymentStatus = "running"
	DeploymentStatusUnhealthy   DeploymentStatus = "unhealthy"
	DeploymentStatusFailed      DeploymentStatus = "failed"
	DeploymentStatusDeploying   DeploymentStatus = "deploying"
)
