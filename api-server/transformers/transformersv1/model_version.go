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
	modelVersionSchemas, err := ToModelVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToModelVersionSchemas")
	}
	modelVersionSchemasMap := make(map[string]*schemasv1.ModelVersionSchema)
	for _, schema := range modelVersionSchemas {
		modelVersionSchemasMap[schema.Uid] = schema
	}
	for _, version := range versions {
		modelSchema, err := GetAssociatedModelSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedModelSchema")
		}
		res = append(res, &schemasv1.ModelVersionFullSchema{
			ModelVersionSchema: *modelVersionSchemasMap[version.Uid],
			Model:              modelSchema,
		})
	}
	return res, nil
}
