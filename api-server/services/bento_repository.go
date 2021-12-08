package services

import (
	"context"
	"fmt"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type bentoRepositoryService struct{}

var BentoRepositoryService = bentoRepositoryService{}

func (s *bentoRepositoryService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.BentoRepository{})
}

type CreateBentoRepositoryOption struct {
	CreatorId      uint
	OrganizationId uint
	Name           string
	Labels         modelschemas.LabelItemsSchema
}

type UpdateBentoRepositoryOption struct {
	Description *string
	Labels      *modelschemas.LabelItemsSchema
}

type ListBentoRepositoryOption struct {
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

func (*bentoRepositoryService) Create(ctx context.Context, opt CreateBentoRepositoryOption) (*models.BentoRepository, error) {
	bentoRepository := models.BentoRepository{
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
	err := mustGetSession(ctx).Create(&bentoRepository).Error
	if err != nil {
		return nil, err
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, opt.CreatorId, opt.OrganizationId, &bentoRepository)
	return &bentoRepository, err
}

func (s *bentoRepositoryService) Update(ctx context.Context, bentoRepository *models.BentoRepository, opt UpdateBentoRepositoryOption) (*models.BentoRepository, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				bentoRepository.Description = *opt.Description
			}
		}()
	}

	if len(updaters) == 0 {
		return bentoRepository, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", bentoRepository.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	if opt.Labels != nil {
		org, err := OrganizationService.GetAssociatedOrganization(ctx, bentoRepository)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, bentoRepository)
		if err != nil {
			return nil, err
		}
	}

	return bentoRepository, err
}

func (s *bentoRepositoryService) Get(ctx context.Context, id uint) (*models.BentoRepository, error) {
	var bentoRepository models.BentoRepository
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bentoRepository).Error
	if err != nil {
		return nil, err
	}
	if bentoRepository.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoRepository, nil
}

func (s *bentoRepositoryService) GetByUid(ctx context.Context, uid string) (*models.BentoRepository, error) {
	var bentoRepository models.BentoRepository
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&bentoRepository).Error
	if err != nil {
		return nil, err
	}
	if bentoRepository.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoRepository, nil
}

func (s *bentoRepositoryService) GetByName(ctx context.Context, organizationId uint, name string) (*models.BentoRepository, error) {
	var bentoRepository models.BentoRepository
	err := getBaseQuery(ctx, s).Where("organization_id = ?", organizationId).Where("name = ?", name).First(&bentoRepository).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bentoRepository %s", name)
	}
	if bentoRepository.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoRepository, nil
}

func (s *bentoRepositoryService) List(ctx context.Context, opt ListBentoRepositoryOption) ([]*models.BentoRepository, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("bento_repository.organization_id = ?", *opt.OrganizationId)
	}
	if opt.CreatorId != nil {
		query = query.Where("bento_repository.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("bento_repository.name in (?)", *opt.Names)
	}
	if opt.Ids != nil {
		query = query.Where("bento_repository.id in (?)", *opt.Ids)
	}
	if opt.CreatorIds != nil {
		query = query.Where("bento_repository.creator_id in (?)", *opt.CreatorIds)
	}
	query = query.Joins("LEFT JOIN bento ON bento.bento_repository_id = bento_repository.id")
	query = query.Joins("LEFT OUTER JOIN bento b2 ON b2.bento_repository_id = bento_repository.id AND bento.id < b2.id")
	query = query.Where("b2.id IS NULL")
	if opt.LastUpdaterIds != nil {
		query = query.Where("bento.creator_id IN (?)", *opt.LastUpdaterIds)
	}
	query = opt.BindQueryWithKeywords(query, "bento_repository")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeBentoRepository)
	query = query.Select("distinct(bento_repository.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bentos := make([]*models.BentoRepository, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("bento_repository.id DESC")
	}
	err = query.Find(&bentos).Error
	if err != nil {
		return nil, 0, err
	}
	return bentos, uint(total), err
}

type IBentoRepositoryAssociate interface {
	GetAssociatedBentoRepositoryId() uint
	GetAssociatedBentoRepositoryCache() *models.BentoRepository
	SetAssociatedBentoRepositoryCache(bentoRepository *models.BentoRepository)
}

func (s *bentoRepositoryService) GetAssociatedBentoRepository(ctx context.Context, associate IBentoRepositoryAssociate) (*models.BentoRepository, error) {
	cache := associate.GetAssociatedBentoRepositoryCache()
	if cache != nil {
		return cache, nil
	}
	bentoRepository, err := s.Get(ctx, associate.GetAssociatedBentoRepositoryId())
	associate.SetAssociatedBentoRepositoryCache(bentoRepository)
	return bentoRepository, err
}

func (s *bentoRepositoryService) GetKubeName(bento *models.BentoRepository) string {
	return fmt.Sprintf("yatai-bento-repository-%s", bento.Name)
}
