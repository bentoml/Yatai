package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type OrganizationSchema struct {
	ResourceSchema
	Creator     *UserSchema `json:"creator"`
	Description string      `json:"description"`
}

type OrganizationFullSchema struct {
	OrganizationSchema
	Config **modelschemas.OrganizationConfigSchema `json:"config"`
}

type OrganizationListSchema struct {
	BaseListSchema
	Items []*OrganizationSchema `json:"items"`
}

type UpdateOrganizationSchema struct {
	Description *string `json:"description"`
}

type CreateOrganizationSchema struct {
	UpdateOrganizationSchema
	Name string `json:"name"`
}
