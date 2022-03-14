package services

import (
	"context"

	"gorm.io/gorm"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/utils"
)

type organizationMemberService struct{}

var OrganizationMemberService = organizationMemberService{}

func (s *organizationMemberService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.OrganizationMember{})
}

type CreateOrganizationMemberOption struct {
	CreatorId      uint
	UserId         uint
	OrganizationId uint
	Role           modelschemas.MemberRole
}

type UpdateOrganizationMemberOption struct {
	Role modelschemas.MemberRole
}

type ListOrganizationMemberOption struct {
	UserId         *uint
	OrganizationId *uint
	Roles          *[]modelschemas.MemberRole
	Order          *string
}

func (s *organizationMemberService) Create(ctx context.Context, operatorId uint, opt CreateOrganizationMemberOption) (*models.OrganizationMember, error) {
	oldMember, err := s.GetBy(ctx, opt.UserId, opt.OrganizationId)
	if err != nil && !utils.IsNotFound(err) {
		return nil, err
	}

	if err == nil {
		return s.Update(ctx, oldMember, operatorId, UpdateOrganizationMemberOption{Role: opt.Role})
	}

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()
	member := &models.OrganizationMember{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		UserAssociate: models.UserAssociate{
			UserId: opt.UserId,
		},
		OrganizationAssociate: models.OrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		Role: opt.Role,
	}
	err = db.Create(member).Error
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *organizationMemberService) Get(ctx context.Context, id uint) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&member).Error
	return &member, err
}

func (s *organizationMemberService) GetBy(ctx context.Context, userId, organizationId uint) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := getBaseQuery(ctx, s).Where("organization_id = ?", organizationId).Where("user_id = ?", userId).First(&member).Error
	return &member, err
}

func (s *organizationMemberService) List(ctx context.Context, opt ListOrganizationMemberOption) ([]*models.OrganizationMember, error) {
	members := make([]*models.OrganizationMember, 0)
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.UserId != nil {
		query = query.Where("user_id = ?", *opt.UserId)
	}
	if opt.Roles != nil {
		query = query.Where("role in (?)", *opt.Roles)
	}
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("id DESC")
	}
	// use Unscoped() to get all members include the soft deleted ones
	err := query.Unscoped().Find(&members).Error
	return members, err
}

func (s *organizationMemberService) ListOrganizationIds(ctx context.Context, userId uint) ([]uint, error) {
	query := s.getBaseDB(ctx)
	query = query.Where("user_id = ?", userId)
	res := make([]uint, 0)
	err := query.Select("organization_id").Find(&res).Error
	return res, err
}

func (s *organizationMemberService) GetResourceType() modelschemas.ResourceType {
	return modelschemas.ResourceTypeOrganization
}

func (s *organizationMemberService) GetOrganization(ctx context.Context, resourceId uint) (*models.Organization, error) {
	return OrganizationService.Get(ctx, resourceId)
}

func (s *organizationMemberService) CheckRoles(ctx context.Context, userId, resourceId uint, roles []modelschemas.MemberRole) (bool, error) {
	q := s.getBaseDB(ctx).
		Where("user_id = ?", userId).
		Where("organization_id = ?", resourceId).
		Where("role in (?)", roles)
	var total int64
	err := q.Count(&total).Error
	return total > 0, err
}

func (s *organizationMemberService) Update(ctx context.Context, m *models.OrganizationMember, operatorId uint, opt UpdateOrganizationMemberOption) (*models.OrganizationMember, error) {
	err := s.getBaseDB(ctx).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"role": opt.Role,
	}).Error
	if err == nil {
		m.Role = opt.Role
	}
	return m, err
}

func (s *organizationMemberService) Delete(ctx context.Context, m *models.OrganizationMember, operatorId uint) (*models.OrganizationMember, error) {
	err := mustGetSession(ctx).Unscoped().Delete(m).Error
	return m, err
}
