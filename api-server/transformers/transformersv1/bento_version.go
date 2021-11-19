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
			Manifest:             version.Manifest,
			BuildAt:              version.BuildAt,
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
	bentoVersionSchemas, err := ToBentoVersionWithBentoSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	bentoVersionSchemasMap := make(map[string]*schemasv1.BentoVersionWithBentoSchema)
	for _, schema := range bentoVersionSchemas {
		bentoVersionSchemasMap[schema.Uid] = schema
	}
	for _, version := range versions {
		modelVersions, _, err := services.ModelVersionService.List(ctx, services.ListModelVersionOption{
			BentoVersionIds: &[]uint{version.ID},
		})
		if err != nil {
			return nil, errors.Wrap(err, "ListModelVersion")
		}
		modelVersionSchemas, err := ToModelVersionSchemas(ctx, modelVersions)
		if err != nil {
			return nil, errors.Wrap(err, "ToModelVersionSchemas")
		}
		res = append(res, &schemasv1.BentoVersionFullSchema{
			BentoVersionWithBentoSchema: *bentoVersionSchemasMap[version.GetUid()],
			ModelVersions:               modelVersionSchemas,
		})
	}
	return res, nil
}

func ToBentoVersionWithBentoSchemas(ctx context.Context, versions []*models.BentoVersion) ([]*schemasv1.BentoVersionWithBentoSchema, error) {
	res := make([]*schemasv1.BentoVersionWithBentoSchema, 0, len(versions))
	bentoVersionSchemas, err := ToBentoVersionSchemas(ctx, versions)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchemas")
	}
	bentoVersionSchemasMap := make(map[string]*schemasv1.BentoVersionSchema)
	for _, schema := range bentoVersionSchemas {
		bentoVersionSchemasMap[schema.Uid] = schema
	}
	bentoIds := make([]uint, 0, len(versions))
	for _, version := range versions {
		bentoIds = append(bentoIds, version.BentoId)
	}
	bentos, _, err := services.BentoService.List(ctx, services.ListBentoOption{
		Ids: &bentoIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListBentos")
	}
	bentoUidToIdMap := make(map[string]uint)
	for _, bento := range bentos {
		bentoUidToIdMap[bento.Uid] = bento.ID
	}
	bentoSchemas, err := ToBentoSchemas(ctx, bentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	bentoSchemasMap := make(map[uint]*schemasv1.BentoSchema)
	for _, schema := range bentoSchemas {
		bentoSchemasMap[bentoUidToIdMap[schema.Uid]] = schema
	}
	for _, version := range versions {
		bentoSchema, ok := bentoSchemasMap[version.BentoId]
		if !ok {
			return nil, errors.Errorf("cannot find bento %d from map", version.BentoId)
		}
		res = append(res, &schemasv1.BentoVersionWithBentoSchema{
			BentoVersionSchema: *bentoVersionSchemasMap[version.GetUid()],
			Bento:              bentoSchema,
		})
	}
	return res, nil
}

type IBentoVersionAssociate interface {
	services.IBentoVersionAssociate
	models.IResource
}

func GetAssociatedBentoVersionFullSchema(ctx context.Context, associate IBentoVersionAssociate) (*schemasv1.BentoVersionFullSchema, error) {
	bentoVersion, err := services.BentoVersionService.GetAssociatedBentoVersion(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	bentoVersionFullSchema, err := ToBentoVersionFullSchema(ctx, bentoVersion)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoVersionSchema")
	}
	return bentoVersionFullSchema, nil
}
