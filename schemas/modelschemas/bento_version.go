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

type BentoVersionManifestApi struct {
	Route  string `json:"route"`
	Doc    string `json:"doc"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type BentoVersionManifestSchema struct {
	Service        string                             `json:"service"`
	BentomlVersion string                             `json:"bentoml_version"`
	Apis           map[string]BentoVersionManifestApi `json:"apis"`
	Models         []string                           `json:"models"`
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
