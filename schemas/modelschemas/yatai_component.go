package modelschemas

type YataiComponentType string

const (
	YataiComponentTypeDeployment YataiComponentType = "deployment"
	YataiComponentTypeLogging    YataiComponentType = "logging"
	YataiComponentTypeMonitoring YataiComponentType = "monitoring"
)

type YataiComponentStatus string

const (
	YataiComponentStatusInstalling YataiComponentStatus = "installing"
	YataiComponentStatusRunning    YataiComponentStatus = "running"
	YataiComponentStatusFailed     YataiComponentStatus = "failed"
)

type YataiComponentInstallerStatus string

const (
	YataiComponentInstallerStatusInstalling YataiComponentInstallerStatus = "installing"
	YataiComponentInstallerStatusRunning    YataiComponentInstallerStatus = "running"
	YataiComponentInstallerStatusFailed     YataiComponentInstallerStatus = "failed"
)
