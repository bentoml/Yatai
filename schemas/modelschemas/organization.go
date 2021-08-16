package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type InfraMinIOSchema struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	BasePath  string `json:"base_path"`
}

type OrganizationConfigSchema struct {
	InfraMinIO *InfraMinIOSchema `json:"infra_minio"`
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
