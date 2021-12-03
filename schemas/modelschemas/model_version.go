package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type ModelVersionUploadStatus string

const (
	ModelVersionUploadStatusPending   ModelVersionUploadStatus = "pending"
	ModelVersionUploadStatusUploading ModelVersionUploadStatus = "uploading"
	ModelVersionUploadStatusSuccess   ModelVersionUploadStatus = "success"
	ModelVersionUploadStatusFailed    ModelVersionUploadStatus = "failed"
)

type ModelVersionImageBuildStatus string

const (
	ModelVersionImageBuildStatusPending  ModelVersionImageBuildStatus = "pending"
	ModelVersionImageBuildStatusBuilding ModelVersionImageBuildStatus = "building"
	ModelVersionImageBuildStatusSuccess  ModelVersionImageBuildStatus = "success"
	ModelVersionImageBuildStatusFailed   ModelVersionImageBuildStatus = "failed"
)

type ModelVersionManifestSchema struct {
	BentomlVersion string                 `json:"bentoml_version"`
	ApiVersion     string                 `json:"api_version"`
	Module         string                 `json:"module"`
	Metadata       map[string]interface{} `json:"metadata"`
	Context        map[string]interface{} `json:"context"`
	Options        map[string]interface{} `json:"options"`
}

func (c *ModelVersionManifestSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *ModelVersionManifestSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
