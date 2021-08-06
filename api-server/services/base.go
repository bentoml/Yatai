package services

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseListOption struct {
	Start  *uint
	Count  *uint
	Search *string
}

func (opt BaseListOption) BindQuery(query *gorm.DB) *gorm.DB {
	if opt.Count != nil {
		query = query.Limit(int(*opt.Count))
	}
	if opt.Start != nil {
		query = query.Offset(int(*opt.Start))
	}
	return query
}

type IDBService interface {
	getBaseDB(ctx context.Context) *gorm.DB
}

func getBaseQuery(ctx context.Context, service IDBService) *gorm.DB {
	return service.getBaseDB(ctx).Preload(clause.Associations)
}
