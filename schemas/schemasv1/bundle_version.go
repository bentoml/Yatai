package schemasv1

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type BundleVersionSchema struct {
	ResourceSchema
	Creator              *UserSchema                            `json:"creator"`
	Version              string                                 `json:"version"`
	Description          string                                 `json:"description"`
	BuildStatus          modelschemas.BundleVersionBuildStatus  `json:"build_status"`
	UploadStatus         modelschemas.BundleVersionUploadStatus `json:"upload_status"`
	UploadStartedAt      *time.Time                             `json:"upload_started_at"`
	UploadFinishedAt     *time.Time                             `json:"upload_finished_at"`
	UploadFinishedReason string                                 `json:"upload_finished_reason"`
	S3Uri                string                                 `json:"s3_uri"`
}

type BundleVersionListSchema struct {
	BaseListSchema
	Items []*BundleVersionSchema `json:"items"`
}

type BundleVersionFullSchema struct {
	BundleVersionSchema
	Bundle *BundleSchema `json:"bundle"`
}

type CreateBundleVersionSchema struct {
	Description string `json:"description"`
	Version     string `json:"version"`
}

type FinishUploadBundleVersionSchema struct {
	Status *modelschemas.BundleVersionUploadStatus `json:"status"`
	Reason *string                                 `json:"reason"`
}
