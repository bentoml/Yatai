package services

import (
	"context"
	"strings"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"
)

type organizationService struct{}

var OrganizationService = organizationService{}

func (*organizationService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Organization{})
}

type CreateOrganizationOption struct {
	CreatorId   uint
	Name        string
	Description *string
}

type UpdateOrganizationOption struct {
	Description *string
}

type ListOrganizationOption struct {
	BaseListOption
	CreatorId *uint
	Ids       *[]uint
}

func (s *organizationService) Create(ctx context.Context, opt CreateOrganizationOption) (*models.Organization, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	org := models.Organization{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
	}
	if opt.Description != nil {
		org.Description = *opt.Description
	}
	err := mustGetSession(ctx).Create(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (s *organizationService) Update(ctx context.Context, o *models.Organization, opt UpdateOrganizationOption) (*models.Organization, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				o.Description = *opt.Description
			}
		}()
	}
	if len(updaters) == 0 {
		return o, nil
	}
	err = s.getBaseDB(ctx).Where("id = ?", o.ID).Updates(updaters).Error
	return o, err
}

func (s *organizationService) Get(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&org).Error
	if err != nil {
		return nil, err
	}
	if org.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &org, nil
}

func (s *organizationService) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization
	err := getBaseQuery(ctx, s).Where("name = ?", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	if org.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &org, nil
}

func (s *organizationService) List(ctx context.Context, opt ListOrganizationOption) ([]*models.Organization, uint, error) {
	orgs := make([]*models.Organization, 0)
	query := getBaseQuery(ctx, s)
	if opt.CreatorId != nil {
		orgIds, err := OrganizationMemberService.ListOrganizationIds(ctx, *opt.CreatorId)
		if err != nil {
			return nil, 0, errors.Wrap(err, "list organization ids")
		}
		// postgresql `in` clause cannot be empty, so push 0 to avoid it empty
		orgIds = append(orgIds, 0)
		query = query.Where("(creator_id = ? or id in (?))", *opt.CreatorId, orgIds)
	}
	if opt.Ids != nil {
		if len(*opt.Ids) == 0 {
			return orgs, 0, nil
		}
		query = query.Where("id in (?)", *opt.Ids)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if opt.Ids == nil {
		query = query.Order("id DESC")
	}
	err = opt.BindQuery(query).Find(&orgs).Error
	return orgs, uint(total), err
}

type IOrganizationAssociate interface {
	GetAssociatedOrganizationId() uint
	GetAssociatedOrganizationCache() *models.Organization
	SetAssociatedOrganizationCache(organization *models.Organization)
}

func (s *organizationService) GetAssociatedOrganization(ctx context.Context, associate IOrganizationAssociate) (*models.Organization, error) {
	cache := associate.GetAssociatedOrganizationCache()
	if cache != nil {
		return cache, nil
	}
	organization, err := s.Get(ctx, associate.GetAssociatedOrganizationId())
	associate.SetAssociatedOrganizationCache(organization)
	return organization, err
}
