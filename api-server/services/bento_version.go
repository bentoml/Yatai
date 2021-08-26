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

type bentoVersionService struct{}

var BentoVersionService = bentoVersionService{}

func (s *bentoVersionService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.BentoVersion{})
}

type CreateBentoVersionOption struct {
	CreatorId   uint
	BentoId     uint
	Version     string
	Description string
}

type UpdateBentoVersionOption struct {
	BuildStatus          *modelschemas.BentoVersionBuildStatus
	UploadStatus         *modelschemas.BentoVersionUploadStatus
	UploadStartedAt      **time.Time
	UploadFinishedAt     **time.Time
	UploadFinishedReason *string
}

type ListBentoVersionOption struct {
	BaseListOption
	BentoId *uint
}

func (*bentoVersionService) Create(ctx context.Context, opt CreateBentoVersionOption) (*models.BentoVersion, error) {
	bentoVersion := models.BentoVersion{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		BentoAssociate: models.BentoAssociate{
			BentoId: opt.BentoId,
		},
		Version:      opt.Version,
		Description:  opt.Description,
		BuildStatus:  modelschemas.BentoVersionBuildStatusPending,
		UploadStatus: modelschemas.BentoVersionUploadStatusPending,
	}
	err := mustGetSession(ctx).Create(&bentoVersion).Error
	if err != nil {
		return nil, err
	}
	return &bentoVersion, err
}

func (s *bentoVersionService) Update(ctx context.Context, b *models.BentoVersion, opt UpdateBentoVersionOption) (*models.BentoVersion, error) {
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

func (s *bentoVersionService) Get(ctx context.Context, id uint) (*models.BentoVersion, error) {
	var bentoVersion models.BentoVersion
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bentoVersion).Error
	if err != nil {
		return nil, err
	}
	if bentoVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoVersion, nil
}

func (s *bentoVersionService) GetByVersion(ctx context.Context, bentoId uint, version string) (*models.BentoVersion, error) {
	var bentoVersion models.BentoVersion
	err := getBaseQuery(ctx, s).Where("bento_id = ?", bentoId).Where("version = ?", version).First(&bentoVersion).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bento version %s", version)
	}
	if bentoVersion.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bentoVersion, nil
}

func (s *bentoVersionService) List(ctx context.Context, opt ListBentoVersionOption) ([]*models.BentoVersion, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.BentoId != nil {
		query = query.Where("bento_id = ?", *opt.BentoId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bentoVersions := make([]*models.BentoVersion, 0)
	query = opt.BindQuery(query).Order("id DESC")
	err = query.Find(&bentoVersions).Error
	if err != nil {
		return nil, 0, err
	}
	return bentoVersions, uint(total), err
}

func (s *bentoVersionService) ListLatestByBentoIds(ctx context.Context, bentoIds []uint) ([]*models.BentoVersion, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select * from bento_version where id in (
					select n.version_id from (
						select bento_id, max(id) as version_id from bento_version 
						where bento_id in (?) group by bento_id
					) as n)`, bentoIds)

	versions := make([]*models.BentoVersion, 0, len(bentoIds))
	err := query.Find(&versions).Error
	if err != nil {
		return nil, err
	}

	return versions, err
}

type IBentoVersionAssociate interface {
	GetAssociatedBentoVersionId() uint
	GetAssociatedBentoVersionCache() *models.BentoVersion
	SetAssociatedBentoVersionCache(version *models.BentoVersion)
}

func (s *bentoVersionService) GetAssociatedBentoVersion(ctx context.Context, associate IBentoVersionAssociate) (*models.BentoVersion, error) {
	cache := associate.GetAssociatedBentoVersionCache()
	if cache != nil {
		return cache, nil
	}
	version, err := s.Get(ctx, associate.GetAssociatedBentoVersionId())
	associate.SetAssociatedBentoVersionCache(version)
	return version, err
}
