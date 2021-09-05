package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type DeploymentSnapshotCanaryRuleType string

const (
	DeploymentSnapshotCanaryRuleTypeWeight DeploymentSnapshotCanaryRuleType = "weight"
	DeploymentSnapshotCanaryRuleTypeHeader DeploymentSnapshotCanaryRuleType = "header"
	DeploymentSnapshotCanaryRuleTypeCookie DeploymentSnapshotCanaryRuleType = "cookie"
)

type DeploymentSnapshotCanaryRule struct {
	Type DeploymentSnapshotCanaryRuleType `json:"type"`

	Weight      *uint   `json:"weight"`
	Header      *string `json:"header"`
	Cookie      *string `json:"cookie"`
	HeaderValue *string `json:"header_value"`
}

type DeploymentSnapshotCanaryRules []*DeploymentSnapshotCanaryRule

func (c *DeploymentSnapshotCanaryRules) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *DeploymentSnapshotCanaryRules) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
