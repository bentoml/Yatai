package services

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type yataiComponentService struct{}

var YataiComponentService = yataiComponentService{}

func (s *yataiComponentService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.YataiComponent{})
}

type CreateYataiComponentOption struct {
	CreatorId      uint
	OrganizationId uint
	ClusterId      uint
	Name           string
	Description    string
	Labels         modelschemas.LabelItemsSchema
	Version        string
	KubeNamespace  string
	Manifest       *modelschemas.YataiComponentManifestSchema
}

type UpdateYataiComponentOption struct {
	Description       *string
	Version           *string
	LatestInstalledAt **time.Time
	LatestHeartbeatAt **time.Time
	Manifest          **modelschemas.YataiComponentManifestSchema
}

func (*yataiComponentService) Create(ctx context.Context, opt CreateYataiComponentOption) (*models.YataiComponent, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	errs = validation.IsDNS1035Label(opt.KubeNamespace)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	now := time.Now()

	yataiComponent := models.YataiComponent{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ClusterAssociate: models.ClusterAssociate{
			ClusterId: opt.ClusterId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		Description:       opt.Description,
		KubeNamespace:     opt.KubeNamespace,
		Manifest:          opt.Manifest,
		Version:           opt.Version,
		LatestInstalledAt: &now,
		LatestHeartbeatAt: &now,
	}
	err := mustGetSession(ctx).Create(&yataiComponent).Error
	if err != nil {
		return nil, err
	}
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return nil, err
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, opt.CreatorId, org.ID, &yataiComponent)
	return &yataiComponent, err
}

func (s *yataiComponentService) Update(ctx context.Context, b *models.YataiComponent, opt UpdateYataiComponentOption) (*models.YataiComponent, error) {
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

	if opt.LatestHeartbeatAt != nil {
		updaters["latest_heartbeat_at"] = *opt.LatestHeartbeatAt
		defer func() {
			if err == nil {
				b.LatestHeartbeatAt = *opt.LatestHeartbeatAt
			}
		}()
	}

	if opt.LatestInstalledAt != nil {
		updaters["latest_installed_at"] = *opt.LatestInstalledAt
		defer func() {
			if err == nil {
				b.LatestInstalledAt = *opt.LatestInstalledAt
			}
		}()
	}

	if opt.Version != nil {
		updaters["version"] = *opt.Version
		defer func() {
			if err == nil {
				b.Version = *opt.Version
			}
		}()
	}

	if opt.Manifest != nil {
		updaters["manifest"] = *opt.Manifest
		defer func() {
			if err == nil {
				b.Manifest = *opt.Manifest
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

func (s *yataiComponentService) Get(ctx context.Context, id uint) (*models.YataiComponent, error) {
	var yataiComponent models.YataiComponent
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&yataiComponent).Error
	if err != nil {
		return nil, err
	}
	if yataiComponent.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &yataiComponent, nil
}

func (s *yataiComponentService) GetByUid(ctx context.Context, uid string) (*models.YataiComponent, error) {
	var yataiComponent models.YataiComponent
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&yataiComponent).Error
	if err != nil {
		return nil, err
	}
	if yataiComponent.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &yataiComponent, nil
}

func (s *yataiComponentService) GetByName(ctx context.Context, clusterId uint, name string) (*models.YataiComponent, error) {
	var yataiComponent models.YataiComponent
	err := getBaseQuery(ctx, s).Where("cluster_id = ?", clusterId).Where("name = ?", name).First(&yataiComponent).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get yataiComponent %s", name)
	}
	if yataiComponent.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &yataiComponent, nil
}

func (s *yataiComponentService) ListByUids(ctx context.Context, uids []string) ([]*models.YataiComponent, error) {
	yataiComponents := make([]*models.YataiComponent, 0, len(uids))
	if len(uids) == 0 {
		return yataiComponents, nil
	}
	err := getBaseQuery(ctx, s).Where("uid in (?)", uids).Find(&yataiComponents).Error
	return yataiComponents, err
}

type ListYataiComponentOption struct {
	Ids            *[]uint `json:"ids"`
	ClusterId      *uint   `json:"cluster_id"`
	OrganizationId *uint   `json:"organization_id"`
}

func (s *yataiComponentService) List(ctx context.Context, opt ListYataiComponentOption) ([]*models.YataiComponent, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.ClusterId != nil {
		query = query.Where("cluster_id = ?", *opt.ClusterId)
	}
	if opt.Ids != nil {
		query = query.Where("id in (?)", *opt.Ids)
	}
	yataiComponents := make([]*models.YataiComponent, 0)
	err := query.Find(&yataiComponents).Error
	if err != nil {
		return nil, err
	}
	return yataiComponents, err
}
