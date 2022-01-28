package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseListOption struct {
	Start             *uint
	Count             *uint
	Search            *string
	Keywords          *[]string
	KeywordFieldNames *[]string
}

func (opt BaseListOption) BindQueryWithLimit(query *gorm.DB) *gorm.DB {
	if opt.Count != nil {
		query = query.Limit(int(*opt.Count))
	}
	if opt.Start != nil {
		query = query.Offset(int(*opt.Start))
	}
	return query
}

func (opt BaseListOption) BindQueryWithKeywords(query *gorm.DB, tableName string) *gorm.DB {
	tableName = query.Statement.Quote(tableName)
	keywordFieldNames := []string{"name"}
	if opt.KeywordFieldNames != nil {
		keywordFieldNames = *opt.KeywordFieldNames
	}
	if opt.Search != nil && *opt.Search != "" {
		sqlPieces := make([]string, 0, len(keywordFieldNames))
		args := make([]interface{}, 0, len(keywordFieldNames))
		for _, keywordFieldName := range keywordFieldNames {
			keywordFieldName = query.Statement.Quote(keywordFieldName)
			sqlPieces = append(sqlPieces, fmt.Sprintf("%s.%s LIKE ?", tableName, keywordFieldName))
			args = append(args, fmt.Sprintf("%%%s%%", *opt.Search))
		}
		query = query.Where(fmt.Sprintf("(%s)", strings.Join(sqlPieces, " OR ")), args...)
	}
	if opt.Keywords != nil {
		sqlPieces := make([]string, 0, len(keywordFieldNames))
		args := make([]interface{}, 0, len(keywordFieldNames))
		for _, keywordFieldName := range keywordFieldNames {
			keywordFieldName = query.Statement.Quote(keywordFieldName)
			sqlPieces_ := make([]string, 0, len(*opt.Keywords))
			for _, keyword := range *opt.Keywords {
				sqlPieces_ = append(sqlPieces_, fmt.Sprintf("%s.%s LIKE ?", tableName, keywordFieldName))
				args = append(args, fmt.Sprintf("%%%s%%", keyword))
			}
			sqlPieces = append(sqlPieces, fmt.Sprintf("(%s)", strings.Join(sqlPieces_, " AND ")))
		}
		query = query.Where(fmt.Sprintf("(%s)", strings.Join(sqlPieces, " OR ")), args...)
	}
	return query
}

type BaseListByLabelsOption struct {
	LabelsList     *[][]modelschemas.LabelItemSchema
	LackLabelsList *[][]modelschemas.LabelItemSchema
}

func (opt BaseListByLabelsOption) BindQueryWithLabels(query *gorm.DB, resourceType modelschemas.ResourceType) *gorm.DB {
	if opt.LabelsList == nil && opt.LackLabelsList == nil {
		return query
	}
	quotedResourceType := query.Statement.Quote(string(resourceType))
	idx := 0
	if opt.LabelsList != nil {
		for _, labels := range *opt.LabelsList {
			alias := fmt.Sprintf("label_%d", idx)
			alias = query.Statement.Quote(alias)
			idx++
			orSqlPieces := make([]string, 0, len(labels))
			orSqlArgs := make([]interface{}, 0, len(labels))
			for _, label := range labels {
				if label.Value != "" {
					orSqlPieces = append(orSqlPieces, fmt.Sprintf("(%s.key = ? AND %s.value = ?)", alias, alias))
					orSqlArgs = append(orSqlArgs, label.Key, label.Value)
				} else {
					orSqlPieces = append(orSqlPieces, fmt.Sprintf("%s.key = ?", alias))
					orSqlArgs = append(orSqlArgs, label.Key)
				}
			}
			query = query.Joins(fmt.Sprintf("JOIN label %s ON %s.resource_type = ? AND %s.resource_id = %s.id AND (%s)", alias, alias, alias, quotedResourceType, strings.Join(orSqlPieces, " OR ")), append([]interface{}{resourceType}, orSqlArgs...)...)
		}
	}
	return query
}

type IDBService interface {
	getBaseDB(ctx context.Context) *gorm.DB
}

func getBaseQuery(ctx context.Context, service IDBService) *gorm.DB {
	return service.getBaseDB(ctx).Preload(clause.Associations)
}
