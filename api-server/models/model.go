package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type Model struct {
	BaseModel
	CreatorAssociate
	ModelRepositoryAssociate
	Version                   string                            `json:"version"`
	Description               string                            `json:"description"`
	UploadStatus              modelschemas.ModelUploadStatus    `json:"upload_status"`
	ImageBuildStatus          modelschemas.ImageBuildStatus     `json:"image_build_status"`
	ImageBuildStatusSyncingAt *time.Time                        `json:"image_build_status_syncing_at"`
	ImageBuildStatusUpdatedAt *time.Time                        `json:"image_build_status_updated_at"`
	UploadStartedAt           *time.Time                        `json:"upload_started_at"`
	UploadFinishedAt          *time.Time                        `json:"upload_finished_at"`
	UploadFinishedReason      string                            `json:"upload_finished_reason"`
	Manifest                  *modelschemas.ModelManifestSchema `json:"manifest" type:"jsonb"`
	BuildAt                   time.Time                         `json:"build_at"`
}

func (b *Model) GetName() string {
	return b.Version
}

func (b *Model) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeModel
}
