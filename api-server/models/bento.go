package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type Bento struct {
	BaseModel
	CreatorAssociate
	BentoRepositoryAssociate
	Version                   string                             `json:"version"`
	Description               string                             `json:"description"`
	FilePath                  string                             `json:"file_path"`
	UploadStatus              modelschemas.BentoUploadStatus     `json:"upload_status"`
	ImageBuildStatus          modelschemas.BentoImageBuildStatus `json:"image_build_status"`
	ImageBuildStatusSyncingAt *time.Time                         `json:"image_build_status_syncing_at"`
	ImageBuildStatusUpdatedAt *time.Time                         `json:"image_build_status_updated_at"`
	UploadStartedAt           *time.Time                         `json:"upload_started_at"`
	UploadFinishedAt          *time.Time                         `json:"upload_finished_at"`
	UploadFinishedReason      string                             `json:"upload_finished_reason"`
	Manifest                  *modelschemas.BentoManifestSchema  `json:"manifest"`
	BuildAt                   time.Time                          `json:"build_at"`
}

func (b *Bento) GetName() string {
	return b.Version
}

func (b *Bento) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeBento
}
