package models

import "github.com/bentoml/yatai-schemas/modelschemas"

type Event struct {
	BaseModel
	NullableOrganizationAssociate
	NullableClusterAssociate
	CreatorAssociate
	Name          string
	Status        modelschemas.EventStatus
	ResourceType  modelschemas.ResourceType
	ResourceId    uint
	OperationName string
	ApiTokenName  string
}
