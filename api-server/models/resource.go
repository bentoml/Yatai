package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type IResource interface {
	IBaseModel
	GetResourceType() modelschemas.ResourceType
	GetName() string
}
