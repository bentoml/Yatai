package models

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type BentoVersion struct {
	BaseModel
	CreatorAssociate
	BentoAssociate
	Version              string                                `json:"version"`
	Description          string                                `json:"description"`
	FilePath             string                                `json:"file_path"`
	UploadStatus         modelschemas.BentoVersionUploadStatus `json:"upload_status"`
	BuildStatus          modelschemas.BentoVersionBuildStatus  `json:"build_status"`
	UploadStartedAt      *time.Time                            `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                            `json:"upload_finished_at"`
	UploadFinishedReason string                                `json:"upload_finished_reason"`
}

func (b *BentoVersion) GetName() string {
	return b.Version
}

func (b *BentoVersion) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeBentoVersion
}
