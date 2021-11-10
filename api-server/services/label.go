package services

import (
	"context"

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
	CreatorId uint
	Resource  models.IResource
	Key       string
	Value     string
}

type ListLabelOption struct {
	BaseListOption
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

func (s *labelService) Get(ctx context.Context, id uint) (*models.Label, error) {
	//TODO
}

func (s *labelService) List(ctx context.Context, opt ListLabelOption) ([]*models.Label, uint, error) {
	//TODO
	var total int64

	return label, uint(total), err
}

func Delete() {
	//TODO
}

func FilterQuery() {
	//TODO
}
