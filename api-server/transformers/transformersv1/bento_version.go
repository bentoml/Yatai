package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/services"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToBentoVersionSchema(ctx context.Context, version *models.BentoVersion) (*schemasv1.BentoVersionSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToBentoVersionSchemas(ctx, []*models.BentoVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	return ss[0], nil
}

func ToBentoVersionSchemas(ctx context.Context, versions []*models.BentoVersion) ([]*schemasv1.BentoVersionSchema, error) {
	res := make([]*schemasv1.BentoVersionSchema, 0, len(versions))
	for _, version := range versions {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		bento, err := services.BentoService.GetAssociatedBento(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBento")
		}
		res = append(res, &schemasv1.BentoVersionSchema{
			ResourceSchema:       ToResourceSchema(version),
			BentoUid:             bento.Uid,
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

func ToBentoVersionFullSchema(ctx context.Context, version *models.BentoVersion) (*schemasv1.BentoVersionFullSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToBentoVersionFullSchemas(ctx, []*models.BentoVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	return ss[0], nil
}

func ToBentoVersionFullSchemas(ctx context.Context, versions []*models.BentoVersion) ([]*schemasv1.BentoVersionFullSchema, error) {
	res := make([]*schemasv1.BentoVersionFullSchema, 0, len(versions))
	bentoVersionSchemas, err := ToBentoVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	bentoVersionSchemasMap := make(map[string]*schemasv1.BentoVersionSchema)
	for _, schema := range bentoVersionSchemas {
		bentoVersionSchemasMap[schema.Uid] = schema
	}
	for _, version := range versions {
		bentoSchema, err := GetAssociatedBentoSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBentoSchema")
		}
		res = append(res, &schemasv1.BentoVersionFullSchema{
			BentoVersionSchema: *bentoVersionSchemasMap[version.GetUid()],
			Bento:              bentoSchema,
		})
	}
	return res, nil
}
