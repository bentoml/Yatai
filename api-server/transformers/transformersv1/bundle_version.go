package transformersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/services"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/pkg/errors"
)

func ToBundleVersionSchema(ctx context.Context, version *models.BundleVersion) (*schemasv1.BundleVersionSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToBundleVersionSchemas(ctx, []*models.BundleVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleVersionSchemas")
	}
	return ss[0], nil
}

func ToBundleVersionSchemas(ctx context.Context, versions []*models.BundleVersion) ([]*schemasv1.BundleVersionSchema, error) {
	res := make([]*schemasv1.BundleVersionSchema, 0, len(versions))
	for _, version := range versions {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		bundle, err := services.BundleService.GetAssociatedBundle(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBundle")
		}
		res = append(res, &schemasv1.BundleVersionSchema{
			ResourceSchema:       ToResourceSchema(version),
			BundleUid:            bundle.Uid,
			Version:              version.Version,
			Creator:              creatorSchema,
			Description:          version.Description,
			BuildStatus:          version.BuildStatus,
			UploadStatus:         version.UploadStatus,
			UploadStartedAt:      version.UploadStartedAt,
			UploadFinishedAt:     version.UploadFinishedAt,
			UploadFinishedReason: version.UploadFinishedReason,
		})
	}
	return res, nil
}

func ToBundleVersionFullSchema(ctx context.Context, version *models.BundleVersion) (*schemasv1.BundleVersionFullSchema, error) {
	if version == nil {
		return nil, nil
	}
	ss, err := ToBundleVersionFullSchemas(ctx, []*models.BundleVersion{version})
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleVersionSchemas")
	}
	return ss[0], nil
}

func ToBundleVersionFullSchemas(ctx context.Context, versions []*models.BundleVersion) ([]*schemasv1.BundleVersionFullSchema, error) {
	res := make([]*schemasv1.BundleVersionFullSchema, 0, len(versions))
	bundleVersionSchemas, err := ToBundleVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBundleVersionSchemas")
	}
	bundleVersionSchemasMap := make(map[string]*schemasv1.BundleVersionSchema)
	for _, schema := range bundleVersionSchemas {
		bundleVersionSchemasMap[schema.Uid] = schema
	}
	for _, version := range versions {
		bundleSchema, err := GetAssociatedBundleSchema(ctx, version)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBundleSchema")
		}
		res = append(res, &schemasv1.BundleVersionFullSchema{
			BundleVersionSchema: *bundleVersionSchemasMap[version.GetUid()],
			Bundle:              bundleSchema,
		})
	}
	return res, nil
}
