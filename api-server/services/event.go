package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type eventService struct{}

func (s *eventService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Event{})
}

var EventService = eventService{}

type CreateEventOption struct {
	Name           string
	OperationName  string
	ApiTokenName   string
	ResourceType   modelschemas.ResourceType
	ResourceId     uint
	CreatorId      uint
	Status         modelschemas.EventStatus
	OrganizationId *uint
	ClusterId      *uint
}

type ListEventOption struct {
	BaseListOption
	OrganizationId *uint
	ClusterId      *uint
	ResourceType   *modelschemas.ResourceType
	ResourceId     *uint
	CreatorId      *uint
	CreatorIds     *[]uint
	Order          *string
	StartedAt      *time.Time
	EndedAt        *time.Time
	OperationNames *[]string
	Status         *modelschemas.EventStatus
}

func (s *eventService) Create(ctx context.Context, opt CreateEventOption) (event *models.Event, err error) {
	// nolint: ineffassign, staticcheck
	db, ctx, df, err := startTransaction(ctx)
	if err != nil {
		return
	}
	defer func() { df(err) }()
	event = &models.Event{
		BaseModel: models.BaseModel{
			Model: gorm.Model{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		NullableOrganizationAssociate: models.NullableOrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		NullableClusterAssociate: models.NullableClusterAssociate{
			ClusterId: opt.ClusterId,
		},
		Name:          opt.Name,
		Status:        opt.Status,
		OperationName: opt.OperationName,
		ResourceType:  opt.ResourceType,
		ResourceId:    opt.ResourceId,
		ApiTokenName:  opt.ApiTokenName,
	}
	err = db.Create(event).Error
	if err != nil {
		return
	}
	return
}

func (s *eventService) ListOperationNames(ctx context.Context, organizationId uint, resourceType modelschemas.ResourceType) (names []string, err error) {
	db := s.getBaseDB(ctx)
	query := db.Raw(`select distinct(operation_name) from event where organization_id = ? and resource_type = ?`, organizationId, resourceType)
	err = query.Find(&names).Error
	return
}

func (s *eventService) List(ctx context.Context, opt ListEventOption) (events []*models.Event, total uint, err error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.ClusterId != nil {
		query = query.Where("cluster_id = ?", *opt.ClusterId)
	}
	if opt.ResourceType != nil {
		query = query.Where("resource_type = ?", *opt.ResourceType)
	}
	if opt.ResourceId != nil {
		query = query.Where("resource_id = ?", *opt.ResourceId)
	}
	if opt.CreatorId != nil {
		query = query.Where("creator_id = ?", *opt.CreatorId)
	}
	if opt.CreatorIds != nil {
		query = query.Where("creator_id in (?)", *opt.CreatorIds)
	}
	if opt.StartedAt != nil {
		query = query.Where("created_at >= ?", *opt.StartedAt)
	}
	if opt.EndedAt != nil {
		query = query.Where("created_at <= ?", *opt.EndedAt)
	}
	if opt.Status != nil {
		query = query.Where("status = ?", *opt.Status)
	}
	if opt.OperationNames != nil {
		query = query.Where("operation_name in (?)", *opt.OperationNames)
	}
	var total_ int64
	err = query.Count(&total_).Error
	if err != nil {
		return
	}
	total = uint(total_)
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query = query.Order("id DESC")
	}
	err = query.Find(&events).Error
	return
}
