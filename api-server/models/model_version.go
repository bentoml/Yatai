package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type ModelVersion struct {
	BaseModel
	CreatorAssociate
	OrganizationAssociate
	ModelAssociate
	Version                   string                                    `json:"version"`
	Description               string                                    `json:"description"`
	UploadStatus              modelschemas.ModelVersionUploadStatus     `json:"upload_status"`
	ImageBuildStatus          modelschemas.ModelVersionImageBuildStatus `json:"image_build_status"`
	ImageBuildStatusSyncingAt *time.Time                                `json:"image_build_status_syncing_at"`
	ImageBuildStatusUpdatedAt *time.Time                                `json:"image_build_status_updated_at"`
	UploadStartedAt           *time.Time                                `json:"upload_started_at"`
	UploadFinishedAt          *time.Time                                `json:"upload_finished_at"`
	UploadFinishedReason      string                                    `json:"upload_finished_reason"`
	Manifest                  *modelschemas.ModelVersionManifestSchema  `json:"manifest"`
	BuildAt                   time.Time                                 `json:"build_at"`
}

func (b *ModelVersion) GetName() string {
	return b.Version
}

func (b *ModelVersion) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeModelVersion
}
