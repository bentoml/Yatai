package services

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type modelRepositoryService struct{}

var ModelRepositoryService = modelRepositoryService{}

func (s *modelRepositoryService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.ModelRepository{})
}

type CreateModelRepositoryOption struct {
	CreatorId      uint
	OrganizationId uint
	Name           string
	Labels         modelschemas.LabelItemsSchema
}

type UpdateModelRepositoryOption struct {
	Description *string
	Labels      *modelschemas.LabelItemsSchema
}

type ListModelRepositoryOption struct {
	BaseListOption
	BaseListByLabelsOption
	OrganizationId *uint
	CreatorId      *uint
	CreatorIds     *[]uint
	LastUpdaterIds *[]uint
	Order          *string
	Names          *[]string
	Ids            *[]uint
}

func (*modelRepositoryService) Create(ctx context.Context, opt CreateModelRepositoryOption) (*models.ModelRepository, error) {
	modelRepository := models.ModelRepository{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
	}
	err := mustGetSession(ctx).Create(&modelRepository).Error
	if err != nil {
		return nil, err
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, opt.CreatorId, opt.OrganizationId, &modelRepository)
	return &modelRepository, err
}

func (s *modelRepositoryService) Update(ctx context.Context, modelRepository *models.ModelRepository, opt UpdateModelRepositoryOption) (*models.ModelRepository, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				modelRepository.Description = *opt.Description
			}
		}()
	}
	if len(updaters) == 0 {
		return modelRepository, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", modelRepository.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}
	if opt.Labels != nil {
		org, err := OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, modelRepository)
		if err != nil {
			return nil, err
		}
	}

	return modelRepository, err
}

func (s *modelRepositoryService) Get(ctx context.Context, id uint) (*models.ModelRepository, error) {
	var modelRepository models.ModelRepository
	err := s.getBaseDB(ctx).Where("id = ?", id).First(&modelRepository).Error
	if err != nil {
		return nil, err
	}
	if modelRepository.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &modelRepository, nil
}

func (s *modelRepositoryService) GetByUid(ctx context.Context, uid string) (*models.ModelRepository, error) {
	var modelRepository models.ModelRepository
	err := s.getBaseDB(ctx).Where("uid = ?", uid).First(&modelRepository).Error
	if err != nil {
		return nil, err
	}
	if modelRepository.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &modelRepository, nil
}

func (s *modelRepositoryService) GetByName(ctx context.Context, organizationId uint, name string) (*models.ModelRepository, error) {
	var modelRepository models.ModelRepository
	err := s.getBaseDB(ctx).Where("organization_id = ?", organizationId).Where("name = ?", name).First(&modelRepository).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get modelRepository %s", name)
	}
	return &modelRepository, nil
}

func (s *modelRepositoryService) List(ctx context.Context, opt ListModelRepositoryOption) ([]*models.ModelRepository, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("model_repository.organization_id = ?", *opt.OrganizationId)
	}
	if opt.CreatorId != nil {
		query = query.Where("model_repository.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("model_repository.name in (?)", *opt.Names)
	}
	if opt.Ids != nil {
		query = query.Where("model_repository.id in (?)", *opt.Ids)
	}
	if opt.CreatorIds != nil {
		query = query.Where("model_repository.creator_id in (?)", *opt.CreatorIds)
	}
	query = query.Joins("LEFT JOIN model ON model.model_repository_id = model_repository.id")
	query = query.Joins("LEFT OUTER JOIN model m2 ON m2.model_repository_id = model_repository.id AND model.id < m2.id")
	query = query.Where("m2.id IS NULL")
	if opt.LastUpdaterIds != nil {
		query = query.Where("model.creator_id IN (?)", *opt.LastUpdaterIds)
	}
	query = opt.BindQueryWithKeywords(query, "model_repository")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeModelRepository)
	query = query.Select("distinct(model_repository.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	modelRepositories := make([]*models.ModelRepository, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("model_repository.id DESC")
	}
	err = query.Find(&modelRepositories).Error
	if err != nil {
		return nil, 0, err
	}
	return modelRepositories, uint(total), nil
}

type IModelRepositoryAssociate interface {
	GetAssociatedModelRepositoryId() uint
	GetAssociatedModelRepositoryCache() *models.ModelRepository
	SetAssociatedModelRepositoryCache(modelRepository *models.ModelRepository)
}

func (s *modelRepositoryService) GetAssociatedModelRepository(ctx context.Context, associate IModelRepositoryAssociate) (*models.ModelRepository, error) {
	cache := associate.GetAssociatedModelRepositoryCache()
	if cache != nil {
		return cache, nil
	}
	modelRepository, err := s.Get(ctx, associate.GetAssociatedModelRepositoryId())
	associate.SetAssociatedModelRepositoryCache(modelRepository)
	return modelRepository, err
}
