package services

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/utils"
)

type cacheService struct{}

var CacheService = cacheService{}

func (*cacheService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Cache{})
}

func (s *cacheService) Set(ctx context.Context, key string, value interface{}) error {
	valueStr, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cache := &models.Cache{
		Key:   key,
		Value: string(valueStr),
	}

	return mustGetSession(ctx).Create(cache).Error
}

func (s *cacheService) Get(ctx context.Context, key string, value interface{}) (exists bool, err error) {
	var cache models.Cache
	err = getBaseQuery(ctx, s).Where("key = ?", key).First(&cache).Error
	if err != nil {
		if utils.IsNotFound(err) {
			err = nil
			return
		}
		return
	}

	exists = true
	err = json.Unmarshal([]byte(cache.Value), value)
	return
}

func (s *cacheService) Delete(ctx context.Context, key string) (exists bool, err error) {
	cache := &models.Cache{}
	err = getBaseQuery(ctx, s).Where("key = ?", key).Unscoped().Delete(cache).Error
	if err != nil {
		if utils.IsNotFound(err) {
			err = nil
			return
		}
		return
	}

	exists = true
	return
}
