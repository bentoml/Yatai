package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToModelSchema(ctx context.Context, model *models.Model) (*schemasv1.ModelSchema, error) {
	if model == nil {
		return nil, nil
	}
	ss, err := ToModelSchemas(ctx, []*models.Model{model})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	return ss[0], nil
}

func ToModelSchemas(ctx context.Context, models_ []*models.Model) ([]*schemasv1.ModelSchema, error) {
	res := make([]*schemasv1.ModelSchema, 0, len(models_))
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, models_)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	for _, model := range models_ {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, model)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		modelRepository, err := services.ModelRepositoryService.GetAssociatedModelRepository(ctx, model)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedModelRepository")
		}
		resourceSchema, ok := resourceSchemasMap[model.GetUid()]
		if !ok {
			return nil, errors.Errorf("resourceSchema not found for model %s", model.GetUid())
		}
		res = append(res, &schemasv1.ModelSchema{
			ResourceSchema:       resourceSchema,
			ModelUid:             modelRepository.Uid,
			Version:              model.Version,
			Creator:              creatorSchema,
			Description:          model.Description,
			ImageBuildStatus:     model.ImageBuildStatus,
			UploadStatus:         model.UploadStatus,
			UploadStartedAt:      model.UploadStartedAt,
			UploadFinishedAt:     model.UploadFinishedAt,
			UploadFinishedReason: model.UploadFinishedReason,
			Manifest:             model.Manifest,
			BuildAt:              model.BuildAt,
		})
	}
	return res, nil
}

func ToModelFullSchema(ctx context.Context, model *models.Model) (*schemasv1.ModelFullSchema, error) {
	if model == nil {
		return nil, nil
	}
	ss, err := ToModelFullSchemas(ctx, []*models.Model{model})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelFullSchemas")
	}
	return ss[0], nil
}

func ToModelFullSchemas(ctx context.Context, models_ []*models.Model) ([]*schemasv1.ModelFullSchema, error) {
	res := make([]*schemasv1.ModelFullSchema, 0, len(models_))
	modelSchemas, err := ToModelWithRepositorySchemas(ctx, models_)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	modelSchemasMap := make(map[string]*schemasv1.ModelWithRepositorySchema)
	for _, schema := range modelSchemas {
		modelSchemasMap[schema.Uid] = schema
	}
	for _, model := range models_ {
		bentos, _, err := services.BentoService.List(ctx, services.ListBentoOption{
			ModelIds: &[]uint{model.ID},
		})
		if err != nil {
			return nil, errors.Wrap(err, "ListBento")
		}
		bentoSchemas, err := ToBentoSchemas(ctx, bentos)
		if err != nil {
			return nil, errors.Wrap(err, "ToBentoSchemas")
		}
		res = append(res, &schemasv1.ModelFullSchema{
			ModelWithRepositorySchema: *modelSchemasMap[model.GetUid()],
			Bentos:                    bentoSchemas,
		})
	}
	return res, nil
}

func ToModelWithRepositorySchemas(ctx context.Context, models_ []*models.Model) ([]*schemasv1.ModelWithRepositorySchema, error) {
	res := make([]*schemasv1.ModelWithRepositorySchema, 0, len(models_))
	modelSchemas, err := ToModelSchemas(ctx, models_)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	modelSchemasMap := make(map[string]*schemasv1.ModelSchema)
	for _, schema := range modelSchemas {
		modelSchemasMap[schema.Uid] = schema
	}
	modelRepositoryIds := make([]uint, 0, len(models_))
	for _, model := range models_ {
		modelRepositoryIds = append(modelRepositoryIds, model.ModelRepositoryId)
	}
	modelRepositories, _, err := services.ModelRepositoryService.List(ctx, services.ListModelRepositoryOption{
		Ids: &modelRepositoryIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListModels")
	}
	modelRepositoryUidToIdMap := make(map[string]uint)
	for _, modelRepository := range modelRepositories {
		modelRepositoryUidToIdMap[modelRepository.Uid] = modelRepository.ID
	}
	modelRepositorySchemas, err := ToModelRepositorySchemas(ctx, modelRepositories)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelRepositorySchemas")
	}
	modelRepositorySchemasMap := make(map[uint]*schemasv1.ModelRepositorySchema)
	for _, schema := range modelRepositorySchemas {
		modelRepositorySchemasMap[modelRepositoryUidToIdMap[schema.Uid]] = schema
	}
	for _, model := range models_ {
		modelRepositorySchema, ok := modelRepositorySchemasMap[model.ModelRepositoryId]
		if !ok {
			return nil, errors.Errorf("cannot find modelRepository %d from map", model.ModelRepositoryId)
		}
		res = append(res, &schemasv1.ModelWithRepositorySchema{
			ModelSchema: *modelSchemasMap[model.GetUid()],
			Repository:  modelRepositorySchema,
		})
	}
	return res, nil
}
