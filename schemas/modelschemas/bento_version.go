package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type BentoVersionUploadStatus string

const (
	BentoVersionUploadStatusPending   BentoVersionUploadStatus = "pending"
	BentoVersionUploadStatusUploading BentoVersionUploadStatus = "uploading"
	BentoVersionUploadStatusSuccess   BentoVersionUploadStatus = "success"
	BentoVersionUploadStatusFailed    BentoVersionUploadStatus = "failed"
)

type BentoVersionImageBuildStatus string

const (
	BentoVersionImageBuildStatusPending  BentoVersionImageBuildStatus = "pending"
	BentoVersionImageBuildStatusBuilding BentoVersionImageBuildStatus = "building"
	BentoVersionImageBuildStatusSuccess  BentoVersionImageBuildStatus = "success"
	BentoVersionImageBuildStatusFailed   BentoVersionImageBuildStatus = "failed"
)

type BentoVersionManifestMetadata struct {
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	ModuleName     string `json:"module_name"`
	ModuleFile     string `json:"module_file"`
}

type BentoVersionManifestApi struct {
	Name       string `json:"name"`
	Docs       string `json:"docs"`
	InputType  string `json:"input_type"`
	OutputType string `json:"output_type"`
}

type BentoVersionManifestArtifact struct {
	Name string `json:"name"`
	Type string `json:"artifact_type"`
}

type BentoVersionManifestSchema struct {
	Metadata BentoVersionManifestMetadata `json:"metadata"`
	Apis     []BentoVersionManifestApi    `json:"apis"`
}

func (c *BentoVersionManifestSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *BentoVersionManifestSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
