package models

import (
	"sync"

	"github.com/lib/pq"

	"github.com/bentoml/yatai-schemas/modelschemas"
)

type TerminalRecord struct {
	BaseModel
	CreatorAssociate
	NullableOrganizationAssociate
	NullableClusterAssociate
	NullableDeploymentAssociate

	ResourceType modelschemas.ResourceType `json:"resource_type"`
	ResourceId   uint                      `json:"resource_id"`

	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`

	Meta    *modelschemas.TerminalRecordMeta `json:"meta"`
	Content pq.StringArray                   `gorm:"type:text[]"`

	Mu sync.Mutex `gorm:"-" json:"-"`
}

func (r *TerminalRecord) GetName() string {
	return r.Uid
}

func (r *TerminalRecord) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeTerminalRecord
}
