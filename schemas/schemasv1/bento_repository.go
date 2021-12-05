package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type BentoRepositorySchema struct {
	ResourceSchema
	Creator      *UserSchema         `json:"creator"`
	Organization *OrganizationSchema `json:"organization"`
	LatestBento  *BentoSchema        `json:"latest_bento"`
	Description  string              `json:"description"`
}

type BentoRepositoryListSchema struct {
	BaseListSchema
	Items []*BentoRepositorySchema `json:"items"`
}

type CreateBentoRepositorySchema struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Labels      modelschemas.LabelItemsSchema `json:"labels"`
}

type UpdateBentoRepositorySchema struct {
	Description *string                        `json:"description"`
	Labels      *modelschemas.LabelItemsSchema `json:"labels,omitempty"`
}
