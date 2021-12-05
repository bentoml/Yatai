package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type BentoUploadStatus string

const (
	BentoUploadStatusPending   BentoUploadStatus = "pending"
	BentoUploadStatusUploading BentoUploadStatus = "uploading"
	BentoUploadStatusSuccess   BentoUploadStatus = "success"
	BentoUploadStatusFailed    BentoUploadStatus = "failed"
)

type BentoImageBuildStatus string

const (
	BentoImageBuildStatusPending  BentoImageBuildStatus = "pending"
	BentoImageBuildStatusBuilding BentoImageBuildStatus = "building"
	BentoImageBuildStatusSuccess  BentoImageBuildStatus = "success"
	BentoImageBuildStatusFailed   BentoImageBuildStatus = "failed"
)

type BentoManifestApi struct {
	Route  string `json:"route"`
	Doc    string `json:"doc"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type BentoManifestSchema struct {
	Service        string                      `json:"service"`
	BentomlVersion string                      `json:"bentoml_version"`
	Apis           map[string]BentoManifestApi `json:"apis"`
	Models         []string                    `json:"models"`
}

func (c *BentoManifestSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *BentoManifestSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
