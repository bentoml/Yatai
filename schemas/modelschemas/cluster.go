package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type ClusterConfigSchema struct {
	IngressIp string `json:"ingress_ip"`
}

func (c *ClusterConfigSchema) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *ClusterConfigSchema) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
