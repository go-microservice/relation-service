package cache

//go:generate mockgen -source=internal/cache/user_follower_cache.go -destination=internal/mock/user_follower_cache_mock.go  -package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-eagle/eagle/pkg/cache"
	"github.com/go-eagle/eagle/pkg/encoding"
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/go-eagle/eagle/pkg/redis"

	"github.com/go-microservice/relation-service/internal/model"
)

const (
	// PrefixUserFollowerCacheKey cache prefix
	PrefixUserFollowerCacheKey = "userFollower:%d"
)

// UserFollower define cache interface
type UserFollowerCache interface {
	SetUserFollowerCache(ctx context.Context, id int64, data *model.UserFollowerModel, duration time.Duration) error
	GetUserFollowerCache(ctx context.Context, id int64) (data *model.UserFollowerModel, err error)
	MultiGetUserFollowerCache(ctx context.Context, ids []int64) (map[string]*model.UserFollowerModel, error)
	MultiSetUserFollowerCache(ctx context.Context, data []*model.UserFollowerModel, duration time.Duration) error
	DelUserFollowerCache(ctx context.Context, id int64) error
}

// userFollowerCache define cache struct
type userFollowerCache struct {
	cache cache.Cache
}

// NewUserFollowerCache new a cache
func NewUserFollowerCache() *userFollowerCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""
	return &userFollowerCache{
		cache: cache.NewRedisCache(redis.RedisClient, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UserFollowerModel{}
		}),
	}
}

// GetUserFollowerCacheKey get cache key
func (c *userFollowerCache) GetUserFollowerCacheKey(id int64) string {
	return fmt.Sprintf(PrefixUserFollowerCacheKey, id)
}

// SetUserFollowerCache write to cache
func (c *userFollowerCache) SetUserFollowerCache(ctx context.Context, id int64, data *model.UserFollowerModel, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUserFollowerCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// GetUserFollowerCache get from cache
func (c *userFollowerCache) GetUserFollowerCache(ctx context.Context, id int64) (data *model.UserFollowerModel, err error) {
	cacheKey := c.GetUserFollowerCacheKey(id)
	err = c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		log.WithContext(ctx).Warnf("get err from redis, err: %+v", err)
		return nil, err
	}
	return data, nil
}

// MultiGetUserFollowerCache batch get cache
func (c *userFollowerCache) MultiGetUserFollowerCache(ctx context.Context, ids []int64) (map[string]*model.UserFollowerModel, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUserFollowerCacheKey(v)
		keys = append(keys, cacheKey)
	}

	// NOTE: 需要在这里make实例化，如果在返回参数里直接定义会报 nil map
	retMap := make(map[string]*model.UserFollowerModel)
	err := c.cache.MultiGet(ctx, keys, retMap)
	if err != nil {
		return nil, err
	}
	return retMap, nil
}

// MultiSetUserFollowerCache batch set cache
func (c *userFollowerCache) MultiSetUserFollowerCache(ctx context.Context, data []*model.UserFollowerModel, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUserFollowerCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}
	return nil
}

// DelUserFollowerCache delete cache
func (c *userFollowerCache) DelUserFollowerCache(ctx context.Context, id int64) error {
	cacheKey := c.GetUserFollowerCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// DelUserFollowerCache set empty cache
func (c *userFollowerCache) SetCacheWithNotFound(ctx context.Context, id int64) error {
	cacheKey := c.GetUserFollowerCacheKey(id)
	err := c.cache.SetCacheWithNotFound(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}
