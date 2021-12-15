package models

import "github.com/bentoml/yatai/schemas/modelschemas"

type LabelConfiguration struct {
	BaseModel
	OrganizationAssociate
	CreatorAssociate

	Key  string                               `json:"key"`
	Info *modelschemas.LabelConfigurationInfo `json:"info"`
}
