package schemasv1

import "github.com/bentoml/yatai/schemas/modelschemas"

type ResourceSchema struct {
	BaseSchema
	Name         string                        `json:"name"`
	ResourceType modelschemas.ResourceType     `json:"resource_type" enum:"user,organization,cluster,bento_repository,bento,deployment,deployment_revision,model_repository,model,api_token"`
	Labels       modelschemas.LabelItemsSchema `json:"labels"`
}

func (s *ResourceSchema) TypeName() string {
	return string(s.ResourceType)
}
