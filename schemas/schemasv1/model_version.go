package schemasv1

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type ModelVersionSchema struct {
	ResourceSchema
	ModelUid             string                                    `json:"model_uid"`
	Creator              *UserSchema                               `json:"creator"`
	Version              string                                    `json:"version"`
	Description          string                                    `json:"description"`
	ImageBuildStatus     modelschemas.ModelVersionImageBuildStatus `json:"image_build_status"`
	UploadStatus         modelschemas.ModelVersionUploadStatus     `json:"upload_status"`
	UploadStartedAt      *time.Time                                `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                                `json:"upload_finished_at"`
	UploadFinishedReason string                                    `json:"upload_finished_reason"`
}

type ModelVersionListSchema struct {
	BaseListSchema
	Items []*ModelVersionSchema `json:"item"`
}

type ModelVersionFullSchema struct {
	ModelVersionSchema
	Model *ModelSchema `json:"model"`
}

type CreateModelVersionSchema struct {
	Version     string                                     `json:"version"`
	Description string                                     `json:"description"`
	Manifest    *modelschemas.ModelVersionManifestSchema   `json:"manifest"`
	BuildAt     string                                     `json:"build_at"`
	Labels      modelschemas.CreateLabelsForResourceSchema `json:"labels"`
}

type UpdateModelVersionSchema struct {
	Description string                                      `json:"description,omitempty"`
	Manifest    *modelschemas.ModelVersionManifestSchema    `json:"manifest"`
	BuildAt     string                                      `json:"build_at"`
	Labels      *modelschemas.CreateLabelsForResourceSchema `json:"labels,omitempty"`
}

type FinishUploadModelVersionSchema struct {
	Status *modelschemas.ModelVersionUploadStatus `json:"status"`
	Reason *string                                `json:"reason"`
}
