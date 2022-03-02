package schemasv1

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type BentoSchema struct {
	ResourceSchema
	BentoRepositoryUid   string                            `json:"bento_repository_uid"`
	Creator              *UserSchema                       `json:"creator"`
	Version              string                            `json:"version"`
	Description          string                            `json:"description"`
	ImageName            string                            `json:"image_name"`
	InClusterImageName   string                            `json:"in_cluster_image_name"`
	ImageBuildStatus     modelschemas.ImageBuildStatus     `json:"image_build_status"`
	UploadStatus         modelschemas.BentoUploadStatus    `json:"upload_status"`
	UploadStartedAt      *time.Time                        `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                        `json:"upload_finished_at"`
	UploadFinishedReason string                            `json:"upload_finished_reason"`
	PresignedUploadUrl   string                            `json:"presigned_upload_url"`
	PresignedDownloadUrl string                            `json:"presigned_download_url"`
	Manifest             *modelschemas.BentoManifestSchema `json:"manifest"`
	BuildAt              time.Time                         `json:"build_at"`
}

type BentoListSchema struct {
	BaseListSchema
	Items []*BentoSchema `json:"items"`
}

type BentoWithRepositorySchema struct {
	BentoSchema
	Repository *BentoRepositorySchema `json:"repository"`
}

type BentoWithRepositoryListSchema struct {
	BaseListSchema
	Items []*BentoWithRepositorySchema `json:"items"`
}

type BentoFullSchema struct {
	BentoWithRepositorySchema
	Models []*ModelSchema `json:"models"`
}

type CreateBentoSchema struct {
	Description string                            `json:"description"`
	Version     string                            `json:"version"`
	Manifest    *modelschemas.BentoManifestSchema `json:"manifest"`
	BuildAt     string                            `json:"build_at"`
	Labels      modelschemas.LabelItemsSchema     `json:"labels"`
}

type UpdateBentoSchema struct {
	Description string                            `json:"description"`
	Version     string                            `json:"version"`
	Manifest    *modelschemas.BentoManifestSchema `json:"manifest"`
	BuildAt     string                            `json:"build_at"`
	Labels      *modelschemas.LabelItemsSchema    `json:"labels,omitempty"`
}

type FinishUploadBentoSchema struct {
	Status *modelschemas.BentoUploadStatus `json:"status"`
	Reason *string                         `json:"reason"`
}
