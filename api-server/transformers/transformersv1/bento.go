package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToBentoSchema(ctx context.Context, bento *models.Bento) (*schemasv1.BentoSchema, error) {
	if bento == nil {
		return nil, nil
	}
	ss, err := ToBentoSchemas(ctx, []*models.Bento{bento})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	return ss[0], nil
}

func ToBentoSchemas(ctx context.Context, bentos []*models.Bento) ([]*schemasv1.BentoSchema, error) {
	res := make([]*schemasv1.BentoSchema, 0, len(bentos))
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, bentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}
	for _, bento := range bentos {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, bento)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedBentoRepository")
		}
		resourceSchema, ok := resourceSchemasMap[bento.GetUid()]
		if !ok {
			return nil, errors.Errorf("resourceSchema not found for bento %s", bento.GetUid())
		}
		imageName, err := services.BentoService.GetImageName(ctx, bento, false)
		if err != nil {
			return nil, errors.Wrap(err, "GetImageName")
		}
		inClusterImageName, err := services.BentoService.GetImageName(ctx, bento, true)
		if err != nil {
			return nil, errors.Wrap(err, "GetInClusterImageName")
		}
		res = append(res, &schemasv1.BentoSchema{
			ResourceSchema:       resourceSchema,
			BentoRepositoryUid:   bentoRepository.Uid,
			Version:              bento.Version,
			Creator:              creatorSchema,
			Description:          bento.Description,
			ImageName:            imageName,
			InClusterImageName:   inClusterImageName,
			ImageBuildStatus:     bento.ImageBuildStatus,
			UploadStatus:         bento.UploadStatus,
			UploadStartedAt:      bento.UploadStartedAt,
			UploadFinishedAt:     bento.UploadFinishedAt,
			UploadFinishedReason: bento.UploadFinishedReason,
			Manifest:             bento.Manifest,
			BuildAt:              bento.BuildAt,
		})
	}
	return res, nil
}

func ToBentoFullSchema(ctx context.Context, bento *models.Bento) (*schemasv1.BentoFullSchema, error) {
	if bento == nil {
		return nil, nil
	}
	ss, err := ToBentoFullSchemas(ctx, []*models.Bento{bento})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoFullSchemas")
	}
	return ss[0], nil
}

func ToBentoFullSchemas(ctx context.Context, bentos []*models.Bento) ([]*schemasv1.BentoFullSchema, error) {
	res := make([]*schemasv1.BentoFullSchema, 0, len(bentos))
	bentoSchemas, err := ToBentoWithRepositorySchemas(ctx, bentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	bentoSchemasMap := make(map[string]*schemasv1.BentoWithRepositorySchema)
	for _, schema := range bentoSchemas {
		bentoSchemasMap[schema.Uid] = schema
	}
	for _, bento := range bentos {
		models_, _, err := services.ModelService.List(ctx, services.ListModelOption{
			BentoIds: &[]uint{bento.ID},
		})
		if err != nil {
			return nil, errors.Wrap(err, "list models")
		}
		modelSchemas, err := ToModelSchemas(ctx, models_)
		if err != nil {
			return nil, errors.Wrap(err, "ToModelSchemas")
		}
		res = append(res, &schemasv1.BentoFullSchema{
			BentoWithRepositorySchema: *bentoSchemasMap[bento.GetUid()],
			Models:                    modelSchemas,
		})
	}
	return res, nil
}

func ToBentoWithRepositorySchemas(ctx context.Context, bentos []*models.Bento) ([]*schemasv1.BentoWithRepositorySchema, error) {
	res := make([]*schemasv1.BentoWithRepositorySchema, 0, len(bentos))
	bentoSchemas, err := ToBentoSchemas(ctx, bentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	bentoSchemasMap := make(map[string]*schemasv1.BentoSchema)
	for _, schema := range bentoSchemas {
		bentoSchemasMap[schema.Uid] = schema
	}
	bentoRepositoryIds := make([]uint, 0, len(bentos))
	for _, bento := range bentos {
		bentoRepositoryIds = append(bentoRepositoryIds, bento.BentoRepositoryId)
	}
	bentoRepositories, _, err := services.BentoRepositoryService.List(ctx, services.ListBentoRepositoryOption{
		Ids: &bentoRepositoryIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListBentoRepositories")
	}
	bentoRepositoryUidToIdMap := make(map[string]uint)
	for _, bentoRepository := range bentoRepositories {
		bentoRepositoryUidToIdMap[bentoRepository.Uid] = bentoRepository.ID
	}
	bentoRepositorySchemas, err := ToBentoRepositorySchemas(ctx, bentoRepositories)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoRepositorySchemas")
	}
	bentoRepositorySchemasMap := make(map[uint]*schemasv1.BentoRepositorySchema)
	for _, schema := range bentoRepositorySchemas {
		bentoRepositorySchemasMap[bentoRepositoryUidToIdMap[schema.Uid]] = schema
	}
	for _, bento := range bentos {
		bentoRepositorySchema, ok := bentoRepositorySchemasMap[bento.BentoRepositoryId]
		if !ok {
			return nil, errors.Errorf("cannot find bentoRepository %d from map", bento.BentoRepositoryId)
		}
		res = append(res, &schemasv1.BentoWithRepositorySchema{
			BentoSchema: *bentoSchemasMap[bento.GetUid()],
			Repository:  bentoRepositorySchema,
		})
	}
	return res, nil
}

type IBentoAssociate interface {
	services.IBentoAssociate
	models.IResource
}

func GetAssociatedBentoFullSchema(ctx context.Context, associate IBentoAssociate) (*schemasv1.BentoFullSchema, error) {
	bento, err := services.BentoService.GetAssociatedBento(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	bentoFullSchema, err := ToBentoFullSchema(ctx, bento)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoFullSchema")
	}
	return bentoFullSchema, nil
}
