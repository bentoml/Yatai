package schemasv1

type BundleSchema struct {
	ResourceSchema
	Creator       *UserSchema          `json:"creator"`
	Organization  *OrganizationSchema  `json:"organization"`
	LatestVersion *BundleVersionSchema `json:"latest_version"`
	Description   string               `json:"description"`
}

type BundleListSchema struct {
	BaseListSchema
	Items []*BundleSchema `json:"items"`
}

type CreateBundleSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateBundleSchema struct {
	Description *string `json:"description"`
}
