package modelschemas

type BentoVersionUploadStatus string

const (
	BentoVersionUploadStatusPending   BentoVersionUploadStatus = "pending"
	BentoVersionUploadStatusUploading BentoVersionUploadStatus = "uploading"
	BentoVersionUploadStatusSuccess   BentoVersionUploadStatus = "success"
	BentoVersionUploadStatusFailed    BentoVersionUploadStatus = "failed"
)

type BentoVersionBuildStatus string

const (
	BentoVersionBuildStatusPending  BentoVersionBuildStatus = "pending"
	BentoVersionBuildStatusBuilding BentoVersionBuildStatus = "building"
	BentoVersionBuildStatusSuccess  BentoVersionBuildStatus = "success"
	BentoVersionBuildStatusFailed   BentoVersionBuildStatus = "failed"
)
