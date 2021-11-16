package modelschemas

type DeploymentStatus string

const (
	DeploymentStatusUnknown     DeploymentStatus = "unknown"
	DeploymentStatusNonDeployed DeploymentStatus = "non-deployed"
	DeploymentStatusRunning     DeploymentStatus = "running"
	DeploymentStatusUnhealthy   DeploymentStatus = "unhealthy"
	DeploymentStatusFailed      DeploymentStatus = "failed"
	DeploymentStatusDeploying   DeploymentStatus = "deploying"
	DeploymentStatusTerminating DeploymentStatus = "terminating"
	DeploymentStatusTerminated  DeploymentStatus = "terminated"
)

func (d DeploymentStatus) Ptr() *DeploymentStatus {
	return &d
}
