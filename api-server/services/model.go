package services

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type modelService struct{}

var ModelService = modelService{}

func (s *modelService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Model{})
}

type CreateModelOption struct {
	CreatorId      uint
	OrganizationId uint
	Name           string
}

type UpdateModelOption struct {
	Description *string
}

type ListModelOption struct {
	BaseListOption
	OrganizationId *uint
}

func (*modelService) Create(ctx context.Context, opt CreateModelOption) (*models.Model, error) {
	model := models.Model{
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
	err := mustGetSession(ctx).Create(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *modelService) Update(ctx context.Context, m *models.Model, opt UpdateModelOption) (*models.Model, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				m.Description = *opt.Description
			}
		}()
	}
	if len(updaters) == 0 {
		return m, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", m.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}
	return m, err
}

func (s *modelService) Get(ctx context.Context, id uint) (*models.Model, error) {
	var model models.Model
	err := s.getBaseDB(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &model, nil
}

func (s *modelService) GetByName(ctx context.Context, organizationId uint, name string) (*models.Model, error) {
	var model models.Model
	err := s.getBaseDB(ctx).Where("organization_id = ?", organizationId).Where("name = ?", name).First(&model).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get model %s", name)
	}
	if model.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &model, nil
}

func (s *modelService) List(ctx context.Context, opt ListModelOption) ([]*models.Model, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	models := make([]*models.Model, 0)
	query = opt.BindQueryWithLimit(query)
	query = query.Order("id DESC")
	err = query.Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, uint(total), nil
}

type IModelAssociate interface {
	GetAssociatedModelId() uint
	GetAssociatedModelCache() *models.Model
	SetAssociatedModelCache(model *models.Model)
}

func (s *modelService) GetAssociatedModel(ctx context.Context, associate IModelAssociate) (*models.Model, error) {
	cache := associate.GetAssociatedModelCache()
	if cache != nil {
		return cache, nil
	}
	model, err := s.Get(ctx, associate.GetAssociatedModelId())
	associate.SetAssociatedModelCache(model)
	return model, err
}
