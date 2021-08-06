package modelschemas

type BundleVersionUploadStatus string

const (
	BundleVersionUploadStatusPending   BundleVersionUploadStatus = "pending"
	BundleVersionUploadStatusUploading BundleVersionUploadStatus = "uploading"
	BundleVersionUploadStatusSuccess   BundleVersionUploadStatus = "success"
	BundleVersionUploadStatusFailed    BundleVersionUploadStatus = "failed"
)

type BundleVersionBuildStatus string

const (
	BundleVersionBuildStatusPending  BundleVersionBuildStatus = "pending"
	BundleVersionBuildStatusBuilding BundleVersionBuildStatus = "building"
	BundleVersionBuildStatusSuccess  BundleVersionBuildStatus = "success"
	BundleVersionBuildStatusFailed   BundleVersionBuildStatus = "failed"
)
