package schemasv1

type ModelSchema struct {
	ResourceSchema
	Creator       *UserSchema         `json:"creator"`
	Organization  *OrganizationSchema `json:"organization"`
	LatestVersion *ModelVersionSchema `json:"latest_version"`
	Description   string              `json:"description"`
}

type ModelListSchema struct {
	BaseListSchema
	Items []*ModelSchema `json:"items"`
}

type CreateModelSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreateLabelSchema
}

type UpdateModelSchema struct {
	Description *string `json:"description"`
	UpdateLabelSchema
}
