package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToModelRepositorySchema(ctx context.Context, modelRepository *models.ModelRepository) (*schemasv1.ModelRepositorySchema, error) {
	if modelRepository == nil {
		return nil, nil
	}
	ss, err := ToModelRepositorySchemas(ctx, []*models.ModelRepository{modelRepository})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelRepositorySchemas")
	}
	return ss[0], nil
}

func ToModelRepositorySchemas(ctx context.Context, modelRepositories []*models.ModelRepository) ([]*schemasv1.ModelRepositorySchema, error) {
	modelRepositoryIds := make([]uint, 0, len(modelRepositories))
	for _, modelRepository := range modelRepositories {
		modelRepositoryIds = append(modelRepositoryIds, modelRepository.ID)
	}
	models_, err := services.ModelService.ListLatestByModelRepositoryIds(ctx, modelRepositoryIds)
	if err != nil {
		return nil, errors.Wrap(err, "List latest models by modelRepository ids")
	}
	modelSchemas, err := ToModelSchemas(ctx, models_)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	modelSchemasMap := make(map[string]*schemasv1.ModelSchema, len(models_))
	for _, s := range modelSchemas {
		modelSchemasMap[s.ModelUid] = s
	}
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, modelRepositories)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	res := make([]*schemasv1.ModelRepositorySchema, 0, len(modelRepositories))
	for _, modelRepository := range modelRepositories {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, modelRepository)
		if err != nil {
			return nil, errors.Wrap(err, "get associated creator schema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, modelRepository)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedOrganizationSchema")
		}
		resourceSchema, ok := resourceSchemasMap[modelRepository.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for modelRepository %s", modelRepository.GetUid())
		}
		res = append(res, &schemasv1.ModelRepositorySchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    modelRepository.Description,
			LatestModel:    modelSchemasMap[modelRepository.GetUid()],
		})
	}
	return res, nil
}

type IModelRepositoryAssociate interface {
	services.IModelRepositoryAssociate
	models.IResource
}

func GetAssociatedModelRepositorySchema(ctx context.Context, associate IModelRepositoryAssociate) (*schemasv1.ModelRepositorySchema, error) {
	user, err := services.ModelRepositoryService.GetAssociatedModelRepository(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	modelRepositorySchema, err := ToModelRepositorySchema(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelRepositorySchema")
	}
	return modelRepositorySchema, nil
}
