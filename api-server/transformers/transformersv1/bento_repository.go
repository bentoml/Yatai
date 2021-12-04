package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

func ToBentoRepositorySchema(ctx context.Context, bentoRepository *models.BentoRepository) (*schemasv1.BentoRepositorySchema, error) {
	if bentoRepository == nil {
		return nil, nil
	}
	ss, err := ToBentoRepositorySchemas(ctx, []*models.BentoRepository{bentoRepository})
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoRepositorySchemas")
	}
	return ss[0], nil
}

func ToBentoRepositorySchemas(ctx context.Context, bentoRepositories []*models.BentoRepository) ([]*schemasv1.BentoRepositorySchema, error) {
	bentoRepositoryIds := make([]uint, 0, len(bentoRepositories))
	for _, bentoRepository := range bentoRepositories {
		bentoRepositoryIds = append(bentoRepositoryIds, bentoRepository.ID)
	}

	bentos, err := services.BentoService.ListLatestByBentoRepositoryIds(ctx, bentoRepositoryIds)
	if err != nil {
		return nil, errors.Wrap(err, "list latest bentos by bentoRepository ids")
	}
	bentoSchemas, err := ToBentoSchemas(ctx, bentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	bentoSchemasMap := make(map[string]*schemasv1.BentoSchema, len(bentos))
	for _, s := range bentoSchemas {
		bentoSchemasMap[s.BentoRepositoryUid] = s
	}

	resourceSchemasMap, err := ToResourceSchemasMap(ctx, bentoRepositories)
	if err != nil {
		return nil, errors.Wrap(err, "ToResourceSchemasMap")
	}

	res := make([]*schemasv1.BentoRepositorySchema, 0, len(bentoRepositories))
	for _, bentoRepository := range bentoRepositories {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, bentoRepository)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedCreatorSchema")
		}
		organizationSchema, err := GetAssociatedOrganizationSchema(ctx, bentoRepository)
		if err != nil {
			return nil, errors.Wrap(err, "GetAssociatedClusterSchema")
		}
		resourceSchema, ok := resourceSchemasMap[bentoRepository.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for bentoRepository %s", bentoRepository.GetUid())
		}
		res = append(res, &schemasv1.BentoRepositorySchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    bentoRepository.Description,
			LatestBento:    bentoSchemasMap[bentoRepository.GetUid()],
		})
	}
	return res, nil
}

type IBentoRepositoryAssociate interface {
	services.IBentoRepositoryAssociate
	models.IResource
}

func GetAssociatedBentoRepositorySchema(ctx context.Context, associate IBentoRepositoryAssociate) (*schemasv1.BentoRepositorySchema, error) {
	bentoRepository, err := services.BentoRepositoryService.GetAssociatedBentoRepository(ctx, associate)
	if err != nil {
		return nil, errors.Wrapf(err, "get %s %s associated cluster", associate.GetResourceType(), associate.GetName())
	}
	bentoRepositorySchema, err := ToBentoRepositorySchema(ctx, bentoRepository)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoRepositorySchema")
	}
	return bentoRepositorySchema, nil
}
