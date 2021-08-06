package services

import (
	"context"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type clusterMemberService struct{}

var ClusterMemberService = clusterMemberService{}

func (s *clusterMemberService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.ClusterMember{})
}

func (s *clusterMemberService) GetResourceType() models.ResourceType {
	return models.ResourceTypeCluster
}

type CreateClusterMemberOption struct {
	CreatorId uint
	UserId    uint
	ClusterId uint
	Role      modelschemas.MemberRole
}

type UpdateClusterMemberOption struct {
	Role modelschemas.MemberRole
}

type ListClusterMemberOption struct {
	UserId    *uint
	ClusterId *uint
	Roles     *[]modelschemas.MemberRole
}

func (s *clusterMemberService) Create(ctx context.Context, operatorId uint, opt CreateClusterMemberOption) (*models.ClusterMember, error) {
	err := MemberService.CanOperate(ctx, &ClusterMemberService, operatorId, opt.ClusterId)
	if err != nil {
		return nil, err
	}

	oldMember, err := s.GetBy(ctx, opt.CreatorId, opt.ClusterId)
	if err != nil && !utils.IsNotFound(err) {
		return nil, err
	}

	if err == nil {
		return s.Update(ctx, oldMember, operatorId, UpdateClusterMemberOption{opt.Role})
	}
	err = nil

	// nolint: ineffassign,staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()
	member := &models.ClusterMember{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		UserAssociate: models.UserAssociate{
			UserId: opt.UserId,
		},
		ClusterAssociate: models.ClusterAssociate{
			ClusterId: opt.ClusterId,
		},
		Role: opt.Role,
	}
	err = db.Create(member).Error
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *clusterMemberService) Update(ctx context.Context, m *models.ClusterMember, operatorId uint, opt UpdateClusterMemberOption) (*models.ClusterMember, error) {
	err := MemberService.CanOperate(ctx, s, operatorId, m.ClusterId)
	if err != nil {
		return nil, err
	}
	err = s.getBaseDB(ctx).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"role": opt.Role,
	}).Error
	if err == nil {
		m.Role = opt.Role
	}
	return m, err
}

func (s *clusterMemberService) Get(ctx context.Context, id uint) (*models.ClusterMember, error) {
	var member models.ClusterMember
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&member).Error
	return &member, err
}

func (s *clusterMemberService) GetBy(ctx context.Context, userId, clusterId uint) (*models.ClusterMember, error) {
	var member models.ClusterMember
	err := getBaseQuery(ctx, s).Where("cluster_id = ?", clusterId).Where("user_id = ?", userId).First(&member).Error
	return &member, err
}

func (s *clusterMemberService) List(ctx context.Context, opt ListClusterMemberOption) ([]*models.ClusterMember, error) {
	members := make([]*models.ClusterMember, 0)
	query := getBaseQuery(ctx, s)
	if opt.ClusterId != nil {
		query = query.Where("cluster_id = ?", *opt.ClusterId)
	}
	if opt.UserId != nil {
		query = query.Where("user_id = ?", *opt.UserId)
	}
	if opt.Roles != nil {
		query = query.Where("role in (?)", *opt.Roles)
	}
	err := query.Order("id DESC").Find(&members).Error
	return members, err
}

func (s *organizationMemberService) ListClusterIds(ctx context.Context, userId uint) ([]uint, error) {
	query := s.getBaseDB(ctx)
	query = query.Where("user_id = ?", userId)
	res := make([]uint, 0)
	err := query.Select("cluster_id").Find(&res).Error
	return res, err
}

func (s *clusterMemberService) GetOrganization(ctx context.Context, resourceId uint) (*models.Organization, error) {
	cluster, err := ClusterService.Get(ctx, resourceId)
	if err != nil {
		return nil, errors.Wrap(err, "get organization")
	}
	return OrganizationService.GetAssociatedOrganization(ctx, cluster)
}

func (s *clusterMemberService) CheckRoles(ctx context.Context, userId, resourceId uint, roles []modelschemas.MemberRole) (bool, error) {
	q := s.getBaseDB(ctx).
		Where("user_id = ?", userId).
		Where("cluster_id = ?", resourceId).
		Where("role in (?)", roles)
	var total int64
	err := q.Count(&total).Error
	return total > 0, err
}

func (s *clusterMemberService) Delete(ctx context.Context, m *models.ClusterMember, operatorId uint) (*models.ClusterMember, error) {
	err := MemberService.CanOperate(ctx, &ClusterMemberService, operatorId, m.ClusterId)
	if err != nil {
		return nil, err
	}
	err = s.getBaseDB(ctx).Unscoped().Delete(m).Error
	return m, err
}
