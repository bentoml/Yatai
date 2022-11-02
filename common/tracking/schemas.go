package tracking

import (
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type YataiEventType string

const (
	YataiDeploymentEvent YataiEventType = "yatai_deployment_event"
	YataiBentoEvent      YataiEventType = "yatai_bento_event"
	YataiModelEvent      YataiEventType = "yatai_model_event"
	YataiLifeCycleEvent  YataiEventType = "yatai_lifecycle_event"
)

type CommonProperties struct {
	EventType       string    `json:"event_type"`
	OrganizationUID string    `json:"organization_uid"`
	Timestamp       time.Time `json:"timestamp"`
	YataiVersion    string    `json:"yatai_version"`
}

func NewCommonProperties(eventType YataiEventType, organizationUID string, yataiVersion string) CommonProperties {
	return CommonProperties{
		EventType:       string(eventType),
		OrganizationUID: organizationUID,
		YataiVersion:    yataiVersion,
		Timestamp:       time.Now(),
	}
}

// type LifecycleEvent struct {
// 	LifecycleEventType string        `json:"lifecycle_eventtype"`
// 	Uptime             time.Duration `json:"uptime"`
// }

type DeploymentEventType string

const (
	DeploymentEventTypeCreate    DeploymentEventType = "create"
	DeploymentEventTypeUpdate    DeploymentEventType = "update"
	DeploymentEventTypeTerminate DeploymentEventType = "terminate"
	DeploymentEventTypeDelete    DeploymentEventType = "delete"
)

type DeploymentEvent struct {
	CommonProperties
	UserUID               string                                              `json:"user_uid"`
	ClusterUID            string                                              `json:"cluster_uid"`
	DeploymentUID         string                                              `json:"deployment_uid"`
	DeploymentEventType   DeploymentEventType                                 `json:"deployment_eventtype"`
	DeploymentStatus      modelschemas.DeploymentStatus                       `json:"deployment_status"`
	DeploymentRevisionID  string                                              `json:"deployment_revision_id,omitempty"`
	DeploymentTargetTypes []modelschemas.DeploymentTargetType                 `json:"deployment_target_types,omitempty"`
	ApiServerResources    []modelschemas.DeploymentTargetResources            `json:"api_server_resources,omitempty"`
	ApiServerHPAConfig    []modelschemas.DeploymentTargetHPAConf              `json:"api_server_hpa_config,omitempty"`
	RunnerResourcesList   []map[string]modelschemas.DeploymentTargetResources `json:"runner_resources_list,omitempty"`
	RunnerHPAConfigList   []map[string]modelschemas.DeploymentTargetHPAConf   `json:"runner_hpa_config_list,omitempty"`
	// DeploymentTargetCanaryRuleTypes [][]modelschemas.DeploymentTargetCanaryRuleType
}

type BentoEventType string

const (
	BentoEventTypeBentoPull BentoEventType = "bento_pull"
	BentoEventTypeBentoPush BentoEventType = "bento_push"
)

type BentoEvent struct {
	CommonProperties
	UserUID                   string                            `json:"user_uid"`
	BentoEventType            BentoEventType                    `json:"bento_eventtype"`
	BentoRepositoryUID        string                            `json:"bentorepository_uid"`
	BentoVersion              string                            `json:"bento_version"`
	BentoUploadStatus         modelschemas.BentoUploadStatus    `json:"bento_upload_status"`
	BentoUploadFinishedReason string                            `json:"bento_upload_finished_reason"`
	BentoTransmissionStrategy modelschemas.TransmissionStrategy `json:"bento_transmission_strategy"`
	BentoSizeBytes            uint                              `json:"bento_size_bytes"`
	NumModels                 int                               `json:"num_models"`
	NumRunners                int                               `json:"num_runners"`
}

type ModelEventType string

const (
	ModelEventTypeModelPull ModelEventType = "model_pull"
	ModelEventTypeModelPush ModelEventType = "model_push"
)

type ModelEvent struct {
	CommonProperties
	UserUID                   string                            `json:"user_uid"`
	ModelEventType            ModelEventType                    `json:"model_eventtype"`
	ModelUID                  string                            `json:"model_uid"`
	ModelUploadStatus         modelschemas.ModelUploadStatus    `json:"model_upload_status"`
	ModelUploadFinishedReason string                            `json:"model_upload_finished_reason"`
	ModelTransmissionStrategy modelschemas.TransmissionStrategy `json:"model_transmission_strategy"`
	ModelSizeBytes            uint                              `json:"model_size_bytes"`
}
