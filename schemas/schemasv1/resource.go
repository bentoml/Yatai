package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type ResourceSchema struct {
	BaseSchema
	Name         string                    `json:"name"`
	ResourceType modelschemas.ResourceType `json:"resource_type" enum:"user,organization,cluster,bundle,bundle_version,deployment"`
}

func (s *ResourceSchema) TypeName() string {
	return string(s.ResourceType)
}
