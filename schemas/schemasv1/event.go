package schemasv1

import (
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"
)

type EventSchema struct {
	BaseSchema
	Resource      interface{}              `json:"resource,omitempty"`
	Name          string                   `json:"name,omitempty"`
	Status        modelschemas.EventStatus `json:"status,omitempty"`
	OperationName string                   `json:"operation_name,omitempty"`
	Creator       *UserSchema              `json:"creator,omitempty"`
	CreatedAt     time.Time                `json:"created_at,omitempty"`
}

type EventListSchema struct {
	BaseListSchema
	Items []*EventSchema `json:"items"`
}
