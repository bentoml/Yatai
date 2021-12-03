package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToModelVersionSchema(ctx context.Context, version *models.ModelVersion) (*schemasv1.ModelVersionSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToModelVersionSchemas(ctx, []*models.ModelVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionSchemas")
	}
	return ss[0], nil
}

func ToModelVersionSchemas(ctx context.Context, versions []*models.ModelVersion) ([]*schemasv1.ModelVersionSchema, error) {
	res := make([]*schemasv1.ModelVersionSchema, 0, len(versions))
	for _, version := range versions {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		model, err := services.ModelService.GetAssociatedModel(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedModel")
		}
		res = append(res, &schemasv1.ModelVersionSchema{
			ResourceSchema:       ToResourceSchema(version),
			ModelUid:             model.Uid,
			Version:              version.Version,
			Creator:              creatorSchema,
			Description:          version.Description,
			ImageBuildStatus:     version.ImageBuildStatus,
			UploadStatus:         version.UploadStatus,
			UploadStartedAt:      version.UploadStartedAt,
			UploadFinishedAt:     version.UploadFinishedAt,
			UploadFinishedReason: version.UploadFinishedReason,
			Manifest:             version.Manifest,
			BuildAt:              version.BuildAt,
		})
	}
	return res, nil
}

func ToModelVersionFullSchema(ctx context.Context, version *models.ModelVersion) (*schemasv1.ModelVersionFullSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToModelVersionFullSchemas(ctx, []*models.ModelVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionFullSchemas")
	}
	return ss[0], nil
}

func ToModelVersionFullSchemas(ctx context.Context, versions []*models.ModelVersion) ([]*schemasv1.ModelVersionFullSchema, error) {
	res := make([]*schemasv1.ModelVersionFullSchema, 0, len(versions))
	modelVersionSchemas, err := ToModelVersionWithModelSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionSchemas")
	}
	modelVersionSchemasMap := make(map[string]*schemasv1.ModelVersionWithModelSchema)
	for _, schema := range modelVersionSchemas {
		modelVersionSchemasMap[schema.Uid] = schema
	}
	for _, version := range versions {
		bentoVersions, _, err := services.BentoVersionService.List(ctx, services.ListBentoVersionOption{
			ModelVersionIds: &[]uint{version.ID},
		})
		if err != nil {
			return nil, errors.Wrap(err, "ListBentoVersion")
		}
		bentoVersionSchemas, err := ToBentoVersionSchemas(ctx, bentoVersions)
		if err != nil {
			return nil, errors.Wrap(err, "ToBentoVersionSchemas")
		}
		res = append(res, &schemasv1.ModelVersionFullSchema{
			ModelVersionWithModelSchema: *modelVersionSchemasMap[version.GetUid()],
			BentoVersions:               bentoVersionSchemas,
		})
	}
	return res, nil
}

func ToModelVersionWithModelSchemas(ctx context.Context, versions []*models.ModelVersion) ([]*schemasv1.ModelVersionWithModelSchema, error) {
	res := make([]*schemasv1.ModelVersionWithModelSchema, 0, len(versions))
	modelVersionSchemas, err := ToModelVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionSchemas")
	}
	modelVersionSchemasMap := make(map[string]*schemasv1.ModelVersionSchema)
	for _, schema := range modelVersionSchemas {
		modelVersionSchemasMap[schema.Uid] = schema
	}
	modelIds := make([]uint, 0, len(versions))
	for _, version := range versions {
		modelIds = append(modelIds, version.ModelId)
	}
	models_, _, err := services.ModelService.List(ctx, services.ListModelOption{
		Ids: &modelIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListModels")
	}
	modelUidToIdMap := make(map[string]uint)
	for _, model := range models_ {
		modelUidToIdMap[model.Uid] = model.ID
	}
	modelSchemas, err := ToModelSchemas(ctx, models_)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelSchemas")
	}
	modelSchemasMap := make(map[uint]*schemasv1.ModelSchema)
	for _, schema := range modelSchemas {
		modelSchemasMap[modelUidToIdMap[schema.Uid]] = schema
	}
	for _, version := range versions {
		modelSchema, ok := modelSchemasMap[version.ModelId]
		if !ok {
			return nil, errors.Errorf("cannot find model %d from map", version.ModelId)
		}
		res = append(res, &schemasv1.ModelVersionWithModelSchema{
			ModelVersionSchema: *modelVersionSchemasMap[version.GetUid()],
			Model:              modelSchema,
		})
	}
	return res, nil
}
