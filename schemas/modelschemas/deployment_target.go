package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type DeploymentTargetType string

const (
	DeploymentTargetTypeStable DeploymentTargetType = "stable"
	DeploymentTargetTypeCanary DeploymentTargetType = "canary"
)

var DeploymentTargetTypeAddrs = map[DeploymentTargetType]string{
	DeploymentTargetTypeStable: "stb",
	DeploymentTargetTypeCanary: "cnr",
}

type DeploymentTargetResourceItem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu"`
}

type DeploymentTargetResources struct {
	Requests *DeploymentTargetResourceItem `json:"requests"`
	Limits   *DeploymentTargetResourceItem `json:"limits"`
}

type DeploymentTargetHPAConf struct {
	CPU         *int32  `json:"cpu,omitempty"`
	GPU         *int32  `json:"gpu,omitempty"`
	Memory      *string `json:"memory,omitempty"`
	QPS         *int64  `json:"qps,omitempty"`
	MinReplicas *int32  `json:"min_replicas,omitempty"`
	MaxReplicas *int32  `json:"max_replicas,omitempty"`
}

type DeploymentTargetConfig struct {
	Resources *DeploymentTargetResources `json:"resources"`
	HPAConf   *DeploymentTargetHPAConf   `json:"hpa_conf,omitempty"`
	Envs      *[]*LabelItemSchema        `json:"envs,omitempty"`
}

func (c *DeploymentTargetConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *DeploymentTargetConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
