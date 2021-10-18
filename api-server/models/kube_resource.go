package models

type KubeResource struct {
	APIVersion  string            `json:"apiVersion,omitempty"`
	Kind        string            `json:"kind,omitempty"`
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	MatchLabels map[string]string `json:"match_labels"`
}
