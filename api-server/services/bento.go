package services

import (
	"context"
	"fmt"

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
}

type UpdateBentoOption struct {
	Description *string
}

type ListBentoOption struct {
	BaseListOption
	OrganizationId *uint
	Names          *[]string
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
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.Names != nil {
		query = query.Where("name in (?)", *opt.Names)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bentos := make([]*models.Bento, 0)
	query = opt.BindQuery(query)
	query = query.Order("id DESC")
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
