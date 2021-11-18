package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type labelService struct{}

var LabelService = labelService{}

func (s *labelService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Label{})
}

type CreateLabelOption struct {
	OrganizationId *uint
	CreatorId      uint
	Resource       models.IResource
	Key            string
	Value          string
}

type UpdateLabelOption struct {
	Value *string
}

type ListLabelOption struct {
	BaseListOption
	OrganizationId *uint
	CreatorId      *uint
	ResourceType   *string
	Key            *string
	Value          *string
}

func (*labelService) Create(ctx context.Context, opt CreateLabelOption) (*models.Label, error) {
	n := time.Now()

	label := &models.Label{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ResourceType: opt.Resource.GetResourceType(),
		ResourceId:   opt.Resource.GetId(),
		Key:          opt.Key,
		Value:        opt.Value,
	}

	label.CreatedAt = n

	err := mustGetSession(ctx).Create(label).Error
	if err != nil {
		return nil, err
	}

	return label, nil
}

func (s *labelService) Update(ctx context.Context, b *models.Label, opt UpdateLabelOption) (*models.Label, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Value != nil {
		updaters["value"] = *opt.Value
		defer func() {
			if err == nil {
				b.Value = *opt.Value
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

func (s *labelService) Get(ctx context.Context, id uint) (*models.Label, error) {
	var label models.Label
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&label).Error
	if err != nil {
		return nil, err
	}
	if label.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &label, nil
}

func (s *labelService) GetByUid(ctx context.Context, uid string) (*models.Label, error) {
	var label models.Label
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&label).Error
	if err != nil {
		return nil, err
	}
	if label.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &label, nil
}

func (s *labelService) Delete(ctx context.Context, label *models.Label) (*models.Label, error) {
	return label, s.getBaseDB(ctx).Unscoped().Delete(label).Error
}

func (s *labelService) List(ctx context.Context, opt ListLabelOption) ([]*models.Label, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.CreatorId != nil {
		query = query.Where("creator_id = ?", *opt.CreatorId)
	}

	if opt.ResourceType != nil {
		query = query.Where("resource_type = ?", *opt.ResourceType)
	}
	if opt.Key != nil {
		query = query.Where("key = ?", *opt.Key)
	}
	if opt.Value != nil {
		query = query.Where("value = ?", *opt.Value)
	}

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	labels := make([]*models.Label, 0)
	query = opt.BindQueryWithLimit(query)
	query = query.Order("id DESC")
	err = query.Find(&labels).Error
	if err != nil {
		return nil, 0, err
	}

	return labels, uint(total), err
}

func Filter() {
	//TODO
}
