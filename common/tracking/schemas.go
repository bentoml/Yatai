package tracking

import (
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type CommonProperties struct {
	OrganizationUID string    `json:"organization_uid"`
	Timestamp       time.Time `json:"timestamp"`
	YataiVersion    string    `json:"yatai_version"`
}

type TriggerEvent struct {
	UserUID string `json:"user_uid"`
}

type LifecycleEvent struct {
	LifecycleEventType string        `json:"lifecycle_eventtype"`
	Uptime             time.Duration `json:"uptime"`
}

type DeploymentEventType string

const (
	DeploymentEventTypeCreate    DeploymentEventType = "create"
	DeploymentEventTypeUpdate    DeploymentEventType = "update"
	DeploymentEventTypeTerminate DeploymentEventType = "terminate"
	DeploymentEventTypeDelete    DeploymentEventType = "delete"
)

type DeploymentEvent struct {
	CommonProperties
	TriggerEvent
	ClusterUID            string                              `json:"cluster_uid"`
	DeploymentUID         string                              `json:"deployment_uid"`
	DeploymentEventType   DeploymentEventType                 `json:"deployment_eventtype"`
	DeploymentStatus      modelschemas.DeploymentStatus       `json:"deployment_status"`
	DeploymentRevisionID  string                              `json:"deployment_revision_id,omitempty"`
	DeploymentTargetTypes []modelschemas.DeploymentTargetType `json:"deployment_target_types,omitempty"`
	// DeploymentTargetCanaryRuleTypes [][]modelschemas.DeploymentTargetCanaryRuleType
	ApiServerResources  []modelschemas.DeploymentTargetResources            `json:"api_server_resources,omitempty"`
	ApiServerHPAConfig  []modelschemas.DeploymentTargetHPAConf              `json:"api_server_hpa_config,omitempty"`
	RunnerResourcesList []map[string]modelschemas.DeploymentTargetResources `json:"runner_resources_list,omitempty"`
	RunnerHPAConfigList []map[string]modelschemas.DeploymentTargetHPAConf   `json:"runner_hpa_config_list,omitempty"`
}

type BentoEventType string

const (
	BentoEventTypeBentoPull BentoEventType = "bento_pull"
	BentoEventTypeBentoPush BentoEventType = "bento_push"
)

type BentoEvent struct {
	CommonProperties
	TriggerEvent
	BentoEventType       BentoEventType                 `json:"bento_eventtype"`
	BentoRepositoryUID   string                         `json:"bentorepository_uid"`
	BentoVersion         string                         `json:"bento_version"`
	UploadStatus         modelschemas.BentoUploadStatus `json:"upload_status"`
	UploadFinishedReason string                         `json:"upload_finished_reason"`
	BentoSizeBytes       uint                           `json:"bento_size_bytes"`
	NumModels            int                            `json:"num_models"`
	NumRunners           int                            `json:"num_runners"`
}
