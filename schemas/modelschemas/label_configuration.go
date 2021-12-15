package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type LabelConfigurationInfo struct {
	Color string `json:"color"`
}

func (c *LabelConfigurationInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *LabelConfigurationInfo) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
