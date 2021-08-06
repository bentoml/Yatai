package models

type DeploymentResourceItem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu"`
}

type DeploymentResources struct {
	Requests *DeploymentResourceItem `json:"requests"`
	Limits   *DeploymentResourceItem `json:"limits"`
}

type DeploymentHPAConf struct {
	CPU         *int32  `json:"cpu,omitempty"`
	GPU         *int32  `json:"gpu,omitempty"`
	Memory      *string `json:"memory,omitempty"`
	QPS         *int64  `json:"qps,omitempty"`
	MaxReplicas *int32  `json:"max_replicas,omitempty"`
}

type DeploymentSnapshot struct {
	BaseModel
	CreatorAssociate
	DeploymentAssociate
	BundleVersionAssociate
}

type DeploymentSnapshotConfig struct {
	Resources *DeploymentResources `json:"resources"`
	HPAConf   *DeploymentHPAConf   `json:"hpa_conf,omitempty"`
}
