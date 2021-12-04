package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type ModelUploadStatus string

const (
	ModelUploadStatusPending   ModelUploadStatus = "pending"
	ModelUploadStatusUploading ModelUploadStatus = "uploading"
	ModelUploadStatusSuccess   ModelUploadStatus = "success"
	ModelUploadStatusFailed    ModelUploadStatus = "failed"
)

type ModelImageBuildStatus string

const (
	ModelImageBuildStatusPending  ModelImageBuildStatus = "pending"
	ModelImageBuildStatusBuilding ModelImageBuildStatus = "building"
	ModelImageBuildStatusSuccess  ModelImageBuildStatus = "success"
	ModelImageBuildStatusFailed   ModelImageBuildStatus = "failed"
)

type ModelManifestSchema struct {
	BentomlVersion string                 `json:"bentoml_version"`
	ApiVersion     string                 `json:"api_version"`
	Module         string                 `json:"module"`
	Metadata       map[string]interface{} `json:"metadata"`
	Context        map[string]interface{} `json:"context"`
	Options        map[string]interface{} `json:"options"`
}

func (c *ModelManifestSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *ModelManifestSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
