package services

import (
	"context"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
	"github.com/pkg/errors"
)

type labelService struct{}

var LabelService = labelService{}

func (s *labelService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Label{})
}

type CreateLabelOption struct {
	OrganizationId uint
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
	ResourceId     *string
}

type ListLabelKeysOption struct {
	OrganizationId *uint
	ResourceType   *modelschemas.ResourceType
	ResourceId     *uint
}

type ListLabelValuesByKeyOption struct {
	OrganizationId *uint
	ResourceType   *modelschemas.ResourceType
	ResourceId     *uint
}

func (*labelService) Create(ctx context.Context, opt CreateLabelOption) (*models.Label, error) {
	label := models.Label{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		ResourceType: opt.Resource.GetResourceType(),
		ResourceId:   opt.Resource.GetId(),
		Key:          opt.Key,
		Value:        opt.Value,
	}

	err := mustGetSession(ctx).Create(&label).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
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
	err = s.getBaseDB(ctx).Where("id = ? ", b.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	return b, err
}

func (s *labelService) Get(ctx context.Context, id uint) (*models.Label, error) {
	var label models.Label
	err := getBaseQuery(ctx, s).Where("id = ? ", id).First(&label).Error
	if err != nil {
		return nil, err
	}
	if label.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &label, nil
}

func (s *labelService) GetByKey(ctx context.Context, uid uint, key string) (*models.Label, error) {
	var label models.Label
	err := getBaseQuery(ctx, s).Where("resource_id = ? ", uid).Where("key = ? ", key).First(&label).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get label by key %s", key)
	}
	if label.ResourceId == 0 {
		return nil, consts.ErrNotFound
	}
	return &label, nil
}

// TODO: revisit this after finishing other functionalities.
func (s *labelService) Delete(ctx context.Context, label *models.Label) (*models.Label, error) {
	return label, s.getBaseDB(ctx).Unscoped().Delete(label).Error
}

func (s *labelService) List(ctx context.Context, opt ListLabelOption) ([]*models.Label, uint, error) {
	query := getBaseQuery(ctx, s)

	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ? ", *opt.OrganizationId)
	}
	if opt.CreatorId != nil {
		query = query.Where("creator_id = ? ", *opt.CreatorId)
	}
	if opt.ResourceType != nil {
		query = query.Where("resource_type = ? ", *opt.ResourceType)
	}
	if opt.ResourceId != nil {
		query = query.Where("resource_id = ? ", *opt.ResourceId)
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

func (s *labelService) ListLabelKeys(ctx context.Context, opt ListLabelKeysOption) (keys []string, err error) {
	query := getBaseQuery(ctx, s).Select("DISTINCT key")
	query = query.Where("organization_id = id ?", opt.OrganizationId)

	if opt.ResourceType != nil {
		query = query.Where("resource_type = ?", *opt.ResourceType)
	}
	if opt.ResourceId != nil {
		query = query.Where("resource_id = ?", *opt.ResourceId)
	}
	err = query.Find(&keys).Error
	return
}

func (s *labelService) ListLabelValuesByKey(ctx context.Context, key string, opt ListLabelValuesByKeyOption) (values []string, err error) {
	query := getBaseQuery(ctx, s).Select("DISTINCT value")
	query = query.Where("organization_id = ?", opt.OrganizationId)

	if opt.ResourceType != nil {
		query = query.Where("resource_type = ?", *opt.ResourceType)
	}
	if opt.ResourceId != nil {
		query = query.Where("resource_id = ?", *opt.ResourceId)
	}
	err = query.Find(&values).Error
	return
}

/* TODO:
func Filter() {
}
*/
