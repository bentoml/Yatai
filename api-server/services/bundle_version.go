package services

import (
	"context"
	"time"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type bundleVersionService struct{}

var BundleVersionService = bundleVersionService{}

func (s *bundleVersionService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.BundleVersion{})
}

type CreateBundleVersionOption struct {
	CreatorId   uint
	BundleId    uint
	Version     string
	Description string
}

type UpdateBundleVersionOption struct {
	BuildStatus          *modelschemas.BundleVersionBuildStatus
	UploadStatus         *modelschemas.BundleVersionUploadStatus
	UploadStartedAt      **time.Time
	UploadFinishedAt     **time.Time
	UploadFinishedReason *string
}

type ListBundleVersionOption struct {
	BaseListOption
	BundleId *uint
}

func (*bundleVersionService) Create(ctx context.Context, opt CreateBundleVersionOption) (*models.BundleVersion, error) {
	bundleVersion := models.BundleVersion{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		BundleAssociate: models.BundleAssociate{
			BundleId: opt.BundleId,
		},
		Version:      opt.Version,
		Description:  opt.Description,
		BuildStatus:  modelschemas.BundleVersionBuildStatusPending,
		UploadStatus: modelschemas.BundleVersionUploadStatusPending,
	}
	err := mustGetSession(ctx).Create(&bundleVersion).Error
	if err != nil {
		return nil, err
	}
	return &bundleVersion, err
}

func (s *bundleVersionService) Update(ctx context.Context, b *models.BundleVersion, opt UpdateBundleVersionOption) (*models.BundleVersion, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.BuildStatus != nil {
		updaters["build_status"] = *opt.BuildStatus
		defer func() {
			if err == nil {
				b.BuildStatus = *opt.BuildStatus
			}
		}()
	}
	if opt.UploadStatus != nil {
		updaters["upload_status"] = *opt.UploadStatus
		defer func() {
			if err == nil {
				b.UploadStatus = *opt.UploadStatus
			}
		}()
	}
	if opt.UploadStartedAt != nil {
		updaters["upload_started_at"] = *opt.UploadStartedAt
		defer func() {
			if err == nil {
				b.UploadStartedAt = *opt.UploadStartedAt
			}
		}()
	}
	if opt.UploadFinishedAt != nil {
		updaters["upload_finished_at"] = *opt.UploadFinishedAt
		defer func() {
			if err == nil {
				b.UploadFinishedAt = *opt.UploadFinishedAt
			}
		}()
	}
	if opt.UploadFinishedReason != nil {
		updaters["upload_finished_reason"] = *opt.UploadFinishedReason
		defer func() {
			if err == nil {
				b.UploadFinishedReason = *opt.UploadFinishedReason
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

func (s *bundleVersionService) Get(ctx context.Context, id uint) (*models.BundleVersion, error) {
	var bundleVersion models.BundleVersion
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bundleVersion).Error
	if err != nil {
		return nil, err
	}
	if bundleVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bundleVersion, nil
}

func (s *bundleVersionService) GetByVersion(ctx context.Context, bundleId uint, version string) (*models.BundleVersion, error) {
	var bundleVersion models.BundleVersion
	err := getBaseQuery(ctx, s).Where("bundle_id = ?", bundleId).Where("version = ?", version).First(&bundleVersion).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle version %s", version)
	}
	if bundleVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bundleVersion, nil
}

func (s *bundleVersionService) List(ctx context.Context, opt ListBundleVersionOption) ([]*models.BundleVersion, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.BundleId != nil {
		query = query.Where("bundle_id = ?", *opt.BundleId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bundleVersions := make([]*models.BundleVersion, 0)
	query = opt.BindQuery(query).Order("id DESC")
	err = query.Find(&bundleVersions).Error
	if err != nil {
		return nil, 0, err
	}
	return bundleVersions, uint(total), err
}

func (s *bundleVersionService) ListLatestByBundleIds(ctx context.Context, bundleIds []uint) ([]*models.BundleVersion, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from bundle_version where id in (
					select n.version_id from (
						select bundle_id, max(id) as version_id from bundle_version 
						where bundle_id in (?) group by bundle_id
					) as n)`, bundleIds)

	versions := make([]*models.BundleVersion, 0, len(bundleIds))
	err := query.Find(&versions).Error
	if err != nil {
		return nil, err
	}

	return versions, err
}

type IBundleVersionAssociate interface {
	GetAssociatedBundleVersionId() uint
	GetAssociatedBundleVersionCache() *models.BundleVersion
	SetAssociatedBundleVersionCache(version *models.BundleVersion)
}

func (s *bundleVersionService) GetAssociatedBundleVersion(ctx context.Context, associate IBundleVersionAssociate) (*models.BundleVersion, error) {
	cache := associate.GetAssociatedBundleVersionCache()
	if cache != nil {
		return cache, nil
	}
	version, err := s.Get(ctx, associate.GetAssociatedBundleVersionId())
	associate.SetAssociatedBundleVersionCache(version)
	return version, err
}
