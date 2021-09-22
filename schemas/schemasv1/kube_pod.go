package schemasv1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type KubePodStatusSchema struct {
	Phase     apiv1.PodPhase `json:"phase"`
	Ready     bool           `json:"ready"`
	StartTime *metav1.Time   `json:"start_time"`
	IsOld     bool           `json:"is_old"`
	IsCanary  bool           `json:"is_canary"`
	HostIp    string         `json:"host_ip"`
}

type KubePodSchema struct {
	Name               string                     `json:"name"`
	NodeName           string                     `json:"node_name"`
	DeploymentSnapshot *DeploymentSnapshotSchema  `json:"deployment_snapshot"`
	CommitId           string                     `json:"commit_id"`
	Status             KubePodStatusSchema        `json:"status"`
	RawStatus          apiv1.PodStatus            `json:"raw_status"`
	PodStatus          modelschemas.KubePodStatus `json:"pod_status"`
	Warnings           []apiv1.Event              `json:"warnings"`
}
