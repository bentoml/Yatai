package transformersv1

import (
	"context"
	"fmt"
	"reflect"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToResourceSchemasMap(ctx context.Context, resourcesItf interface{}) (map[string]schemasv1.ResourceSchema, error) {
	resources := make([]models.IResource, 0)
	// nolint: exhaustive
	switch reflect.TypeOf(resourcesItf).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(resourcesItf)

		for i := 0; i < s.Len(); i++ {
			v := s.Index(i).Interface()
			if resource, ok := v.(models.IResource); ok {
				resources = append(resources, resource)
			} else {
				return nil, fmt.Errorf("ToResourceSchemasMap: invalid type %v", reflect.TypeOf(v))
			}
		}
	default:
		return nil, fmt.Errorf("ToResourceSchemasMap: unsupported type %v", reflect.TypeOf(resourcesItf))
	}

	resourceIdsMap := make(map[modelschemas.ResourceType][]uint)
	for _, resource := range resources {
		resourceIdsMap[resource.GetResourceType()] = append(resourceIdsMap[resource.GetResourceType()], resource.GetId())
	}
	labelsMap := make(map[string][]*models.Label)
	genLabelsMapKey := func(resourceType modelschemas.ResourceType, resourceId uint) string {
		return fmt.Sprintf("%s:%d", resourceType, resourceId)
	}
	for resourceType := range resourceIdsMap {
		resourceType := resourceType
		resourceIds := resourceIdsMap[resourceType]
		labels, _, err := services.LabelService.List(ctx, services.ListLabelOption{
			ResourceType: &resourceType,
			ResourceIds:  &resourceIds,
		})
		if err != nil {
			return nil, err
		}
		for _, label := range labels {
			key := genLabelsMapKey(label.ResourceType, label.ResourceId)
			labelsMap[key] = append(labelsMap[key], label)
		}
	}
	resourceSchemas := make([]schemasv1.ResourceSchema, 0, len(resources))
	for _, resource := range resources {
		labels := labelsMap[genLabelsMapKey(resource.GetResourceType(), resource.GetId())]
		labelSchemas := make([]modelschemas.LabelItemSchema, 0, len(labels))
		for _, label := range labels {
			labelSchemas = append(labelSchemas, modelschemas.LabelItemSchema{
				Key:   label.Key,
				Value: label.Value,
			})
		}
		resourceSchema := schemasv1.ResourceSchema{
			BaseSchema:   ToBaseSchema(resource),
			Name:         resource.GetName(),
			ResourceType: resource.GetResourceType(),
			Labels:       labelSchemas,
		}
		resourceSchemas = append(resourceSchemas, resourceSchema)
	}
	resourceSchemasMap := make(map[string]schemasv1.ResourceSchema)
	for _, resourceSchema := range resourceSchemas {
		resourceSchemasMap[resourceSchema.Uid] = resourceSchema
	}
	return resourceSchemasMap, nil
}

func ToResourceSchema(ctx context.Context, resource models.IResource) (schemasv1.ResourceSchema, error) {
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, []models.IResource{resource})
	if err != nil {
		return schemasv1.ResourceSchema{}, err
	}
	return resourceSchemasMap[resource.GetUid()], nil
}
