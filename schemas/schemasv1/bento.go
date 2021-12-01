package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type BentoSchema struct {
	ResourceSchema
	Creator       *UserSchema         `json:"creator"`
	Organization  *OrganizationSchema `json:"organization"`
	LatestVersion *BentoVersionSchema `json:"latest_version"`
	Description   string              `json:"description"`
}

type BentoListSchema struct {
	BaseListSchema
	Items []*BentoSchema `json:"items"`
}

type CreateBentoSchema struct {
	Name        string                                     `json:"name"`
	Description string                                     `json:"description"`
	Labels      modelschemas.CreateLabelsForResourceSchema `json:"labels"`
}

type UpdateBentoSchema struct {
	Description *string                                     `json:"description"`
	Labels      *modelschemas.CreateLabelsForResourceSchema `json:"labels,omitempty"`
}
