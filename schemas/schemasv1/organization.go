package schemasv1

type OrganizationSchema struct {
	ResourceSchema
	Creator     *UserSchema `json:"creator"`
	Description string      `json:"description"`
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
