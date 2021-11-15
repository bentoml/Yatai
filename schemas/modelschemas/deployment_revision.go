package modelschemas

type DeploymentRevisionStatus string

const (
	DeploymentRevisionStatusActive   DeploymentRevisionStatus = "active"
	DeploymentRevisionStatusInactive DeploymentRevisionStatus = "inactive"
)

func DeploymentRevisionStatusPtr(status DeploymentRevisionStatus) *DeploymentRevisionStatus {
	return &status
}
