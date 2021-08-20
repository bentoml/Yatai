package services

import (
	"context"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type deploymentSnapshotService struct{}

var DeploymentSnapshotService = deploymentSnapshotService{}

func (s *deploymentSnapshotService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.DeploymentSnapshot{})
}

type CreateDeploymentSnapshotOption struct {
	CreatorId       uint
	DeploymentId    uint
	BundleVersionId uint
}

type ListDeploymentSnapshotOption struct {
	BaseListOption
	DeploymentId uint
}

func (*deploymentSnapshotService) Create(ctx context.Context, opt CreateDeploymentSnapshotOption) (*models.DeploymentSnapshot, error) {
	snapshot := models.DeploymentSnapshot{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		DeploymentAssociate: models.DeploymentAssociate{
			DeploymentId: opt.DeploymentId,
		},
		BundleVersionAssociate: models.BundleVersionAssociate{
			BundleVersionId: opt.BundleVersionId,
		},
	}
	err := mustGetSession(ctx).Create(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, err
}

func (s *deploymentSnapshotService) Get(ctx context.Context, id uint) (*models.DeploymentSnapshot, error) {
	var snapshot models.DeploymentSnapshot
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	if snapshot.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &snapshot, nil
}

func (s *deploymentSnapshotService) List(ctx context.Context, opt ListDeploymentSnapshotOption) ([]*models.DeploymentSnapshot, uint, error) {
	query := getBaseQuery(ctx, s)
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	deployments := make([]*models.DeploymentSnapshot, 0)
	query = opt.BindQuery(query)
	err = query.Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}
	return deployments, uint(total), err
}

type IDeploymentSnapshotAssociate interface {
	GetAssociatedDeploymentSnapshotId() uint
	GetAssociatedDeploymentSnapshotCache() *models.DeploymentSnapshot
	SetAssociatedDeploymentSnapshotCache(deployment *models.DeploymentSnapshot)
}

func (s *deploymentSnapshotService) GetAssociatedDeploymentSnapshot(ctx context.Context, associate IDeploymentSnapshotAssociate) (*models.DeploymentSnapshot, error) {
	cache := associate.GetAssociatedDeploymentSnapshotCache()
	if cache != nil {
		return cache, nil
	}
	deployment, err := s.Get(ctx, associate.GetAssociatedDeploymentSnapshotId())
	associate.SetAssociatedDeploymentSnapshotCache(deployment)
	return deployment, err
}
