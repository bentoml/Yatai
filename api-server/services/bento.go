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

type bentoService struct{}

var BentoService = bentoService{}

func (s *bentoService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Bento{})
}

type CreateBentoOption struct {
	CreatorId      uint
	OrganizationId uint
	Name           string
	Labels         modelschemas.LabelItemsSchema
}

type UpdateBentoOption struct {
	Description *string
	Labels      *modelschemas.LabelItemsSchema
}

type ListBentoOption struct {
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

func (*bentoService) Create(ctx context.Context, opt CreateBentoOption) (*models.Bento, error) {
	bento := models.Bento{
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
	err := mustGetSession(ctx).Create(&bento).Error
	if err != nil {
		return nil, err
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, opt.CreatorId, opt.OrganizationId, &bento)
	return &bento, err
}

func (s *bentoService) Update(ctx context.Context, b *models.Bento, opt UpdateBentoOption) (*models.Bento, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				b.Description = *opt.Description
			}
		}()
	}

	if len(updaters) == 0 {
		return b, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", b.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	if opt.Labels != nil {
		org, err := OrganizationService.GetAssociatedOrganization(ctx, b)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, b)
		if err != nil {
			return nil, err
		}
	}

	return b, err
}

func (s *bentoService) Get(ctx context.Context, id uint) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bento).Error
	if err != nil {
		return nil, err
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) GetByUid(ctx context.Context, uid string) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&bento).Error
	if err != nil {
		return nil, err
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) GetByName(ctx context.Context, organizationId uint, name string) (*models.Bento, error) {
	var bento models.Bento
	err := getBaseQuery(ctx, s).Where("organization_id = ?", organizationId).Where("name = ?", name).First(&bento).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bento %s", name)
	}
	if bento.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bento, nil
}

func (s *bentoService) List(ctx context.Context, opt ListBentoOption) ([]*models.Bento, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("bento.organization_id = ?", *opt.OrganizationId)
	}
	if opt.CreatorId != nil {
		query = query.Where("bento.creator_id = ?", *opt.CreatorId)
	}
	if opt.Names != nil {
		query = query.Where("bento.name in (?)", *opt.Names)
	}
	if opt.Ids != nil {
		query = query.Where("bento.id in (?)", *opt.Ids)
	}
	if opt.CreatorIds != nil {
		query = query.Where("bento.creator_id in (?)", *opt.CreatorIds)
	}
	query = query.Joins("LEFT JOIN bento_version ON bento_version.bento_id = bento.id")
	query = query.Joins("LEFT OUTER JOIN bento_version v2 ON v2.bento_id = bento.id AND bento_version.id < v2.id")
	query = query.Where("v2.id IS NULL")
	if opt.LastUpdaterIds != nil {
		query = query.Where("bento_version.creator_id IN (?)", *opt.LastUpdaterIds)
	}
	query = opt.BindQueryWithKeywords(query, "bento")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeBento)
	query = query.Select("distinct(bento.*)")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bentos := make([]*models.Bento, 0)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("bento.id DESC")
	}
	err = query.Find(&bentos).Error
	if err != nil {
		return nil, 0, err
	}
	return bentos, uint(total), err
}

type IBentoAssociate interface {
	GetAssociatedBentoId() uint
	GetAssociatedBentoCache() *models.Bento
	SetAssociatedBentoCache(bento *models.Bento)
}

func (s *bentoService) GetAssociatedBento(ctx context.Context, associate IBentoAssociate) (*models.Bento, error) {
	cache := associate.GetAssociatedBentoCache()
	if cache != nil {
		return cache, nil
	}
	bento, err := s.Get(ctx, associate.GetAssociatedBentoId())
	associate.SetAssociatedBentoCache(bento)
	return bento, err
}

func (s *bentoService) GetKubeName(bento *models.Bento) string {
	return fmt.Sprintf("yatai-bento-%s", bento.Name)
}
