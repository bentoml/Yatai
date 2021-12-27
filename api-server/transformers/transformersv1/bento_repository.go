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
	bentoRepositoryUid2BentoSchema := make(map[string]*schemasv1.BentoSchema, len(bentos))
	for _, s := range bentoSchemas {
		bentoRepositoryUid2BentoSchema[s.BentoRepositoryUid] = s
	}

	bentoCountMap, err := services.BentoService.CountByBentoRepositoryIds(ctx, bentoRepositoryIds)
	if err != nil {
		return nil, errors.Wrap(err, "list count bentos by bentoRepository ids")
	}

	deploymentCountMap, err := services.DeploymentService.CountByBentoRepositoryIds(ctx, bentoRepositoryIds)
	if err != nil {
		return nil, errors.Wrap(err, "list count deployments by bentoRepository ids")
	}

	latestBentosMap, err := services.BentoService.GroupByBentoRepositoryIds(ctx, bentoRepositoryIds, 3)
	if err != nil {
		return nil, errors.Wrap(err, "list latest bentos by bentoRepository ids")
	}

	allBentos := make([]*models.Bento, 0)
	for _, bentos := range latestBentosMap {
		allBentos = append(allBentos, bentos...)
	}
	allBentoSchemas, err := ToBentoSchemas(ctx, allBentos)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoSchemas")
	}
	bentoSchemasMap := make(map[string]*schemasv1.BentoSchema, len(allBentoSchemas))
	for _, s := range allBentoSchemas {
		bentoSchemasMap[s.Uid] = s
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
		latestBentos := latestBentosMap[bentoRepository.ID]
		latestBentoSchemas := make([]*schemasv1.BentoSchema, 0, len(latestBentos))
		for _, bento := range latestBentos {
			s, ok := bentoSchemasMap[bento.GetUid()]
			if !ok {
				return nil, errors.Errorf("bento schema not found for bento %s", bento.GetUid())
			}
			latestBentoSchemas = append(latestBentoSchemas, s)
		}
		res = append(res, &schemasv1.BentoRepositorySchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Organization:   organizationSchema,
			Description:    bentoRepository.Description,
			LatestBento:    bentoRepositoryUid2BentoSchema[bentoRepository.GetUid()],
			NBentos:        bentoCountMap[bentoRepository.ID],
			NDeployments:   deploymentCountMap[bentoRepository.ID],
			LatestBentos:   latestBentoSchemas,
		})
	}
	return res, nil
}

func ToBentoRepositoryWithLatestDeploymentsSchemas(ctx context.Context, bentoRepositories []*models.BentoRepository) ([]*schemasv1.BentoRepositoryWithLatestDeploymentsSchema, error) {
	bentoRepositoryIds := make([]uint, 0, len(bentoRepositories))
	for _, bentoRepository := range bentoRepositories {
		bentoRepositoryIds = append(bentoRepositoryIds, bentoRepository.ID)
	}

	latestDeploymentsMap, err := services.DeploymentService.GroupByBentoRepositoryIds(ctx, bentoRepositoryIds, 3)
	if err != nil {
		return nil, errors.Wrap(err, "list latest deployments by bentoRepository ids")
	}
	allDeployments := make([]*models.Deployment, 0)
	for _, deployments := range latestDeploymentsMap {
		allDeployments = append(allDeployments, deployments...)
	}
	allDeploymentSchemas, err := ToDeploymentSchemas(ctx, allDeployments)
	if err != nil {
		return nil, errors.Wrap(err, "ToDeploymentSchemas")
	}
	deploymentSchemasMap := make(map[string]*schemasv1.DeploymentSchema, len(allDeploymentSchemas))
	for _, s := range allDeploymentSchemas {
		deploymentSchemasMap[s.Uid] = s
	}

	schemas, err := ToBentoRepositorySchemas(ctx, bentoRepositories)
	if err != nil {
		return nil, errors.Wrap(err, "ToBentoRepositorySchemas")
	}

	schemasMap := make(map[string]*schemasv1.BentoRepositorySchema, len(schemas))
	for _, s := range schemas {
		schemasMap[s.Uid] = s
	}

	res := make([]*schemasv1.BentoRepositoryWithLatestDeploymentsSchema, 0, len(bentoRepositories))
	for _, bentoRepository := range bentoRepositories {
		latestDeployments := latestDeploymentsMap[bentoRepository.ID]
		latestDeploymentSchemas := make([]*schemasv1.DeploymentSchema, 0, len(latestDeployments))
		for _, deployment := range latestDeployments {
			s, ok := deploymentSchemasMap[deployment.GetUid()]
			if !ok {
				return nil, errors.Errorf("deployment schema not found for deployment %s", deployment.GetUid())
			}
			latestDeploymentSchemas = append(latestDeploymentSchemas, s)
		}
		schema := schemasMap[bentoRepository.GetUid()]
		res = append(res, &schemasv1.BentoRepositoryWithLatestDeploymentsSchema{
			BentoRepositorySchema: *schema,
			LatestDeployments:     latestDeploymentSchemas,
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
