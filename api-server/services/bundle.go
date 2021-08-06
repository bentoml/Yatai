package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"
)

type bundleService struct{}

var BundleService = bundleService{}

func (s *bundleService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Bundle{})
}

type CreateBundleOption struct {
	CreatorId uint
	ClusterId uint
	Name      string
}

type UpdateBundleOption struct {
	Description *string
}

type ListBundleOption struct {
	BaseListOption
	ClusterId *uint
}

func (*bundleService) Create(ctx context.Context, opt CreateBundleOption) (*models.Bundle, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	bundle := models.Bundle{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ClusterAssociate: models.ClusterAssociate{
			ClusterId: opt.ClusterId,
		},
	}
	err := mustGetSession(ctx).Create(&bundle).Error
	if err != nil {
		return nil, err
	}
	return &bundle, err
}

func (s *bundleService) Update(ctx context.Context, b *models.Bundle, opt UpdateBundleOption) (*models.Bundle, error) {
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

func (s *bundleService) Get(ctx context.Context, id uint) (*models.Bundle, error) {
	var bundle models.Bundle
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&bundle).Error
	if err != nil {
		return nil, err
	}
	if bundle.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bundle, nil
}

func (s *bundleService) GetByName(ctx context.Context, clusterId uint, name string) (*models.Bundle, error) {
	var bundle models.Bundle
	err := getBaseQuery(ctx, s).Where("cluster_id = ?", clusterId).Where("name = ?", name).First(&bundle).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle %s", name)
	}
	if bundle.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &bundle, nil
}

func (s *bundleService) List(ctx context.Context, opt ListBundleOption) ([]*models.Bundle, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.ClusterId != nil {
		query = query.Where("cluster_id = ?", *opt.ClusterId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	bundles := make([]*models.Bundle, 0)
	query = opt.BindQuery(query)
	err = query.Find(&bundles).Error
	if err != nil {
		return nil, 0, err
	}
	return bundles, uint(total), err
}

type IBundleAssociate interface {
	GetAssociatedBundleId() uint
	GetAssociatedBundleCache() *models.Bundle
	SetAssociatedBundleCache(bundle *models.Bundle)
}

func (s *bundleService) GetAssociatedBundle(ctx context.Context, associate IBundleAssociate) (*models.Bundle, error) {
	cache := associate.GetAssociatedBundleCache()
	if cache != nil {
		return cache, nil
	}
	bundle, err := s.Get(ctx, associate.GetAssociatedBundleId())
	associate.SetAssociatedBundleCache(bundle)
	return bundle, err
}

func (s *bundleService) GetKubeName(bundle *models.Bundle) string {
	return fmt.Sprintf("yatai-bundle-%s", bundle.Name)
}
