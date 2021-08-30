package schemasv1

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type BentoVersionSchema struct {
	ResourceSchema
	BentoUid             string                                    `json:"bento_uid"`
	Creator              *UserSchema                               `json:"creator"`
	Version              string                                    `json:"version"`
	Description          string                                    `json:"description"`
	ImageBuildStatus     modelschemas.BentoVersionImageBuildStatus `json:"image_build_status"`
	UploadStatus         modelschemas.BentoVersionUploadStatus     `json:"upload_status"`
	UploadStartedAt      *time.Time                                `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                                `json:"upload_finished_at"`
	UploadFinishedReason string                                    `json:"upload_finished_reason"`
	PresignedS3Url       string                                    `json:"presigned_s3_url"`
	Manifest             *modelschemas.BentoVersionManifestSchema  `json:"manifest"`
}

type BentoVersionListSchema struct {
	BaseListSchema
	Items []*BentoVersionSchema `json:"items"`
}

type BentoVersionFullSchema struct {
	BentoVersionSchema
	Bento *BentoSchema `json:"bento"`
}

type CreateBentoVersionSchema struct {
	Description string                                   `json:"description"`
	Version     string                                   `json:"version"`
	Manifest    *modelschemas.BentoVersionManifestSchema `json:"manifest"`
	BuildAt     string                                   `json:"build_at"`
}

type FinishUploadBentoVersionSchema struct {
	Status *modelschemas.BentoVersionUploadStatus `json:"status"`
	Reason *string                                `json:"reason"`
}
