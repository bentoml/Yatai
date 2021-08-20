package transformersv1

import (
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToResourceSchema(resource models.IResource) schemasv1.ResourceSchema {
	return schemasv1.ResourceSchema{
		BaseSchema:   ToBaseSchema(resource),
		Name:         resource.GetName(),
		ResourceType: resource.GetResourceType(),
	}
}
