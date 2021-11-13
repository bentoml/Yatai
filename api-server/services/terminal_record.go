package services

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type terminalRecordService struct{}

var TerminalRecordService = terminalRecordService{}

func (s *terminalRecordService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.TerminalRecord{})
}

type CreateTerminalRecordOption struct {
	CreatorId      uint
	OrganizationId *uint
	ClusterId      *uint
	DeploymentId   *uint
	Resource       models.IResource
	Shell          string
	PodName        string
	ContainerName  string
}

type ListTerminalRecordOption struct {
	BaseListOption
	OrganizationId *uint
	ClusterId      *uint
	DeploymentId   *uint
}

func (s *terminalRecordService) Create(ctx context.Context, opt CreateTerminalRecordOption) (*models.TerminalRecord, error) {
	n := time.Now()
	env := modelschemas.TerminalRecordEnv{
		TERM:  "xterm-256color",
		SHELL: opt.Shell,
	}
	meta := modelschemas.TerminalRecordMeta{
		Version:   2,
		Env:       &env,
		Timestamp: n.Unix(),
	}

	r := &models.TerminalRecord{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		NullableOrganizationAssociate: models.NullableOrganizationAssociate{
			OrganizationId: opt.OrganizationId,
		},
		NullableClusterAssociate: models.NullableClusterAssociate{
			ClusterId: opt.ClusterId,
		},
		NullableDeploymentAssociate: models.NullableDeploymentAssociate{
			DeploymentId: opt.DeploymentId,
		},
		ResourceType:  opt.Resource.GetResourceType(),
		ResourceId:    opt.Resource.GetId(),
		Meta:          &meta,
		PodName:       opt.PodName,
		ContainerName: opt.ContainerName,
	}

	r.CreatedAt = n

	err := mustGetSession(ctx).Create(r).Error
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *terminalRecordService) Get(ctx context.Context, id uint) (*models.TerminalRecord, error) {
	var terminalRecord models.TerminalRecord
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&terminalRecord).Error
	if err != nil {
		return nil, err
	}
	if terminalRecord.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &terminalRecord, nil
}

func (s *terminalRecordService) GetByUid(ctx context.Context, uid string) (*models.TerminalRecord, error) {
	var terminalRecord models.TerminalRecord
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&terminalRecord).Error
	if err != nil {
		return nil, err
	}
	if terminalRecord.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &terminalRecord, nil
}

func (s *terminalRecordService) List(ctx context.Context, opt ListTerminalRecordOption) ([]*models.TerminalRecord, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Where("organization_id = ?", *opt.OrganizationId)
	}
	if opt.ClusterId != nil {
		query = query.Where("cluster_id = ?", *opt.ClusterId)
	}
	if opt.DeploymentId != nil {
		query = query.Where("deployment_id = ?", *opt.DeploymentId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	terminalRecords := make([]*models.TerminalRecord, 0)
	query = opt.BindQueryWithLimit(query)
	query = query.Order("id DESC")
	err = query.Find(&terminalRecords).Error
	if err != nil {
		return nil, 0, err
	}
	return terminalRecords, uint(total), err
}

func (s *terminalRecordService) Append(ctx context.Context, r *models.TerminalRecord, recordType modelschemas.RecordType, value string) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	n := time.Now()
	dt := n.Sub(r.CreatedAt)
	arr := []interface{}{dt.Seconds(), recordType, value}
	c, err := json.Marshal(&arr)
	if err != nil {
		return err
	}
	r.Content = append(r.Content, string(c))

	return nil
}

func (s *terminalRecordService) SaveContent(ctx context.Context, r *models.TerminalRecord) error {
	db := s.getBaseDB(ctx)
	return db.Where("id = ?", r.ID).Updates(&models.TerminalRecord{Content: r.Content, Meta: r.Meta}).Error
}
