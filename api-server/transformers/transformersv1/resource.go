package transformersv1

import (
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

var resourceTypeMapping = map[models.ResourceType]schemasv1.ResourceType{
	models.ResourceTypeUser:          schemasv1.ResourceTypeUser,
	models.ResourceTypeOrganization:  schemasv1.ResourceTypeOrganization,
	models.ResourceTypeCluster:       schemasv1.ResourceTypeCluster,
	models.ResourceTypeBundle:        schemasv1.ResourceTypeBundle,
	models.ResourceTypeBundleVersion: schemasv1.ResourceTypeBundleVersion,
	models.ResourceTypeDeployment:    schemasv1.ResourceTypeDeployment,
}

var reversedResourceTypeMapping = make(map[schemasv1.ResourceType]models.ResourceType, len(resourceTypeMapping))

func init() {
	for k, v := range resourceTypeMapping {
		reversedResourceTypeMapping[v] = k
	}
}

func ToSchemaResourceType(resourceType models.ResourceType) schemasv1.ResourceType {
	return resourceTypeMapping[resourceType]
}

func ToModelResourceType(resourceType schemasv1.ResourceType) models.ResourceType {
	return reversedResourceTypeMapping[resourceType]
}

func ToResourceSchema(resource models.IResource) schemasv1.ResourceSchema {
	return schemasv1.ResourceSchema{
		BaseSchema:   ToBaseSchema(resource),
		Name:         resource.GetName(),
		ResourceType: ToSchemaResourceType(resource.GetResourceType()),
	}
}
