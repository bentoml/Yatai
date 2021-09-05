package modelschemas

import (
	"database/sql/driver"
	"encoding/json"
)

type DeploymentSnapshotStatus string

const (
	DeploymentSnapshotStatusActive   DeploymentSnapshotStatus = "active"
	DeploymentSnapshotStatusInactive DeploymentSnapshotStatus = "inactive"
)

func DeploymentSnapshotStatusPtr(status DeploymentSnapshotStatus) *DeploymentSnapshotStatus {
	return &status
}

type DeploymentSnapshotType string

const (
	DeploymentSnapshotTypeStable DeploymentSnapshotType = "stable"
	DeploymentSnapshotTypeCanary DeploymentSnapshotType = "canary"
)

var DeploymentSnapshotTypeAddrs = map[DeploymentSnapshotType]string{
	DeploymentSnapshotTypeStable: "stb",
	DeploymentSnapshotTypeCanary: "cnr",
}

type DeploymentSnapshotResourceItem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu"`
}

type DeploymentSnapshotResources struct {
	Requests *DeploymentSnapshotResourceItem `json:"requests"`
	Limits   *DeploymentSnapshotResourceItem `json:"limits"`
}

type DeploymentSnapshotHPAConf struct {
	CPU         *int32  `json:"cpu,omitempty"`
	GPU         *int32  `json:"gpu,omitempty"`
	Memory      *string `json:"memory,omitempty"`
	QPS         *int64  `json:"qps,omitempty"`
	MinReplicas *int32  `json:"min_replicas,omitempty"`
	MaxReplicas *int32  `json:"max_replicas,omitempty"`
}

type DeploymentSnapshotConfig struct {
	Resources *DeploymentSnapshotResources `json:"resources"`
	HPAConf   *DeploymentSnapshotHPAConf   `json:"hpa_conf,omitempty"`
}

func (c *DeploymentSnapshotConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal([]byte(value.(string)), c)
}

func (c *DeploymentSnapshotConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
