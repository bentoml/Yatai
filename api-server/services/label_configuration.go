package services

import (
	"context"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type labelConfigurationService struct{}

func (s *labelConfigurationService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.LabelConfiguration{})
}

var LabelConfigurationService = labelConfigurationService{}

type CreateLabelConfigurationOption struct {
	OrganizationId uint
	CreatorId      uint
	Key            string
	Info           *modelschemas.LabelConfigurationInfo
}

type UpdateLabelConfigurationOption struct {
	Info *modelschemas.LabelConfigurationInfo
}

type ListLabelConfigurationOption struct {
	BaseListOption
	OrganizationId uint
	Key            *string
}

func (s *labelConfigurationService) Create(ctx context.Context, opt CreateLabelConfigurationOption) (*models.LabelConfiguration, error) {
	labelConfiguration := &models.LabelConfiguration{
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		Key:  opt.Key,
		Info: opt.Info,
	}
	err := mustGetSession(ctx).Create(labelConfiguration).Error
	return labelConfiguration, err
}

func (s *labelConfigurationService) Update(ctx context.Context, labelConfiguration *models.LabelConfiguration, opt UpdateLabelConfigurationOption) (*models.LabelConfiguration, error) {
	labelConfiguration.Info = opt.Info
	err := s.getBaseDB(ctx).Save(labelConfiguration).Error
	return labelConfiguration, err
}

func (s *labelConfigurationService) GetByKey(ctx context.Context, organizationId uint, key string) (*models.LabelConfiguration, error) {
	labelConfiguration := &models.LabelConfiguration{}
	err := s.getBaseDB(ctx).Where("organization_id = ? AND key = ?", organizationId, key).First(labelConfiguration).Error
	return labelConfiguration, err
}

func (s *labelConfigurationService) List(ctx context.Context, opt ListLabelConfigurationOption) ([]*models.LabelConfiguration, uint, error) {
	q := s.getBaseDB(ctx).Where("organization_id = ?", opt.OrganizationId)
	if opt.Key != nil {
		q = q.Where("key = ?", *opt.Key)
	}
	var total int64
	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	labelConfigurations := make([]*models.LabelConfiguration, 0)
	q = opt.BindQueryWithLimit(q)
	err = q.Find(&labelConfigurations).Error
	return labelConfigurations, uint(total), err
}

func (s *labelConfigurationService) Delete(ctx context.Context, labelConfiguration *models.LabelConfiguration) (*models.LabelConfiguration, error) {
	return labelConfiguration, s.getBaseDB(ctx).Unscoped().Delete(labelConfiguration).Error
}
