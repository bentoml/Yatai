package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type BundleVersion struct {
	BaseModel
	CreatorAssociate
	BundleAssociate
	Version              string                                 `json:"version"`
	Description          string                                 `json:"description"`
	FilePath             string                                 `json:"file_path"`
	UploadStatus         modelschemas.BundleVersionUploadStatus `json:"upload_status"`
	BuildStatus          modelschemas.BundleVersionBuildStatus  `json:"build_status"`
	UploadStartedAt      *time.Time                             `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                             `json:"upload_finished_at"`
	UploadFinishedReason string                                 `json:"upload_finished_reason"`
}

func (b *BundleVersion) GetName() string {
	return b.Version
}

func (b *BundleVersion) GetResourceType() ResourceType {
	return ResourceTypeBundleVersion
}
