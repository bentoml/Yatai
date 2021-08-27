package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type AWSS3Schema struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
}

type AWSECRSchema struct {
	RepositoryURI string `json:"repository_uri"`
	Region        string `json:"region"`
}

type OrganizationConfigAWSSchema struct {
	AccessKeyId     string        `json:"access_key_id"`
	SecretAccessKey string        `json:"secret_access_key"`
	ECR             *AWSECRSchema `json:"ecr"`
	S3              *AWSS3Schema  `json:"s3"`
}

type OrganizationConfigSchema struct {
	MajorClusterUid string                       `json:"major_cluster_uid"`
	AWS             *OrganizationConfigAWSSchema `json:"aws"`
}

func (c *OrganizationConfigSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *OrganizationConfigSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
