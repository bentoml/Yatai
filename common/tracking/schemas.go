package tracking

import (
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type CommonProperties struct {
	OrganisationUID string
	Timestamp       time.Time
	YataiVersion    string
}

type TriggerEvent struct {
	UserUID string
}

type LifecycleEvent struct {
	LifecycleEventType string
	uptime             time.Duration
}

type DeploymentEventType string

const (
	DeploymentEventTypeCreate    DeploymentEventType = "create"
	DeploymentEventTypeStart     DeploymentEventType = "start"
	DeploymentEventTypeUpdate    DeploymentEventType = "update"
	DeploymentEventTypeTerminate DeploymentEventType = "terminate"
	DeploymentEventTypeDelete    DeploymentEventType = "delete"
)

type DeploymentEvent struct {
	CommonProperties
	TriggerEvent
	ClusterUID            string
	DeploymentUID         string
	DeploymentEventType   DeploymentEventType
	DeploymentStatus      modelschemas.DeploymentStatus
	DeploymentRevisionID  string
	DeploymentTargetTypes []modelschemas.DeploymentTargetType
	// DeploymentTargetCanaryRuleTypes [][]modelschemas.DeploymentTargetCanaryRuleType
	ApiServerResources  []modelschemas.DeploymentTargetResources
	ApiServerHPAConfig  []modelschemas.DeploymentTargetHPAConf
	RunnerResourcesList []map[string]modelschemas.DeploymentTargetResources
	RunnerHPAConfigList []map[string]modelschemas.DeploymentTargetHPAConf
}
