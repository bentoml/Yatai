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
	PresignedS3Url       string                                    `json:"presigned_s3_url"`
	Manifest             *modelschemas.ModelVersionManifestSchema  `json:"manifest"`
	BuildAt              time.Time                                 `json:"build_at"`
}

type ModelVersionListSchema struct {
	BaseListSchema
	Items []*ModelVersionSchema `json:"items"`
}

type ModelVersionWithModelSchema struct {
	ModelVersionSchema
	Model *ModelSchema `json:"model"`
}

type ModelVersionWithModelListSchema struct {
	BaseListSchema
	Items []*ModelVersionWithModelSchema `json:"items"`
}

type ModelVersionFullSchema struct {
	ModelVersionWithModelSchema
	BentoVersions []*BentoVersionSchema `json:"bento_versions"`
}

type CreateModelVersionSchema struct {
	Version     string                                   `json:"version"`
	Description string                                   `json:"description"`
	Manifest    *modelschemas.ModelVersionManifestSchema `json:"manifest"`
	BuildAt     string                                   `json:"build_at"`
	Labels      modelschemas.LabelItemsSchema            `json:"labels"`
}

type UpdateModelVersionSchema struct {
	Description string                                   `json:"description,omitempty"`
	Manifest    *modelschemas.ModelVersionManifestSchema `json:"manifest"`
	BuildAt     string                                   `json:"build_at"`
	Labels      *modelschemas.LabelItemsSchema           `json:"labels,omitempty"`
}

type FinishUploadModelVersionSchema struct {
	Status *modelschemas.ModelVersionUploadStatus `json:"status"`
	Reason *string                                `json:"reason"`
}
