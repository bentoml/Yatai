package transformersv1

import (
	"context"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
)

func ToEventSchemas(ctx context.Context, events []*models.Event) ([]*schemasv1.EventSchema, error) {
	creatorIds := make([]uint, 0, len(events))
	for _, event := range events {
		creatorIds = append(creatorIds, event.CreatorId)
	}
	users, err := services.UserService.ListByIds(ctx, creatorIds)
	if err != nil {
		return nil, err
	}
	userUidsMap := make(map[uint]string)
	for _, user := range users {
		userUidsMap[user.ID] = user.Uid
	}
	userSchemas, err := ToUserSchemas(ctx, users)
	if err != nil {
		return nil, err
	}
	userSchemasMap := make(map[string]*schemasv1.UserSchema)
	for _, userSchema := range userSchemas {
		userSchemasMap[userSchema.Uid] = userSchema
	}
	resourceIdsGroup := make(map[modelschemas.ResourceType][]uint)
	for _, event := range events {
		resourceIds, ok := resourceIdsGroup[event.ResourceType]
		if !ok {
			resourceIds = make([]uint, 0)
		}
		resourceIds = append(resourceIds, event.ResourceId)
		resourceIdsGroup[event.ResourceType] = resourceIds
	}
	resourceSchemasGroup := make(map[modelschemas.ResourceType]map[string]interface{})
	resourceUidsGroup := make(map[modelschemas.ResourceType]map[uint]string)
	for resourceType, resourceIds := range resourceIdsGroup {
		resources, err := services.ResourceService.List(ctx, resourceType, resourceIds)
		if err != nil {
			return nil, err
		}
		switch resources_ := resources.(type) {
		case []*models.Bento:
			schemas, err := ToBentoWithRepositorySchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.BentoRepository:
			schemas, err := ToBentoRepositorySchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.Model:
			schemas, err := ToModelWithRepositorySchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.ModelRepository:
			schemas, err := ToModelRepositorySchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.Organization:
			schemas, err := ToOrganizationSchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.Cluster:
			schemas, err := ToClusterSchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.User:
			schemas, err := ToUserSchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		case []*models.Deployment:
			schemas, err := ToDeploymentSchemas(ctx, resources_)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources_ {
				resourceUids, ok := resourceUidsGroup[resourceType]
				if !ok {
					resourceUids = make(map[uint]string)
				}
				resourceUids[resource.ID] = resource.Uid
				resourceUidsGroup[resourceType] = resourceUids
			}
			for _, schema := range schemas {
				if _, ok := resourceSchemasGroup[resourceType]; !ok {
					resourceSchemasGroup[resourceType] = make(map[string]interface{})
				}
				resourceSchemasGroup[resourceType][schema.Uid] = schema
			}
		default:
		}
	}
	eventSchemas := make([]*schemasv1.EventSchema, 0, len(events))
	for _, event := range events {
		userUid, ok := userUidsMap[event.CreatorId]
		var userSchema *schemasv1.UserSchema
		if ok {
			userSchema = userSchemasMap[userUid]
		}
		eventSchema := &schemasv1.EventSchema{
			BaseSchema: schemasv1.BaseSchema{
				Uid:       event.Uid,
				UpdatedAt: event.UpdatedAt,
				CreatedAt: event.CreatedAt,
			},
			Creator:       userSchema,
			Name:          event.Name,
			OperationName: event.OperationName,
			ApiTokenName:  event.ApiTokenName,
			Status:        event.Status,
		}
		resourceUids, ok := resourceUidsGroup[event.ResourceType]
		if ok {
			resourceUid, ok := resourceUids[event.ResourceId]
			if ok {
				resourceSchemas, ok := resourceSchemasGroup[event.ResourceType]
				if ok {
					resourceSchema := resourceSchemas[resourceUid]
					eventSchema.Resource = resourceSchema
				}
			}
		}
		eventSchemas = append(eventSchemas, eventSchema)
	}
	return eventSchemas, nil
}
