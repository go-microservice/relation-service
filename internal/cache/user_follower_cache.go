package cache

//go:generate mockgen -source=internal/cache/user_follower_cache.go -destination=internal/mock/user_follower_cache_mock.go  -package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-eagle/eagle/pkg/cache"
	"github.com/go-eagle/eagle/pkg/encoding"
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/redis/go-redis/v9"

	"github.com/go-microservice/relation-service/internal/model"
)

const (
	// PrefixUserFollowerCacheKey cache prefix
	PrefixUserFollowerCacheKey = "user:follower:%d_%d"
)

// UserFollower define cache interface
type UserFollowerCache interface {
	SetUserFollowerCache(ctx context.Context, userID, followedUID int64, data *model.UserFollowerModel, duration time.Duration) error
	GetUserFollowerCache(ctx context.Context, userID, followedUID int64) (data *model.UserFollowerModel, err error)
	DelUserFollowerCache(ctx context.Context, userID, followedUID int64) error
}

// userFollowerCache define cache struct
type userFollowerCache struct {
	cache cache.Cache
}

// NewUserFollowerCache new a cache
func NewUserFollowerCache(rdb *redis.Client) UserFollowerCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""
	return &userFollowerCache{
		cache: cache.NewRedisCache(rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UserFollowerModel{}
		}),
	}
}

// GetUserFollowerCacheKey get cache key
func (c *userFollowerCache) GetUserFollowerCacheKey(userID, followedUID int64) string {
	return fmt.Sprintf(PrefixUserFollowerCacheKey, userID, followedUID)
}

// SetUserFollowerCache write to cache
func (c *userFollowerCache) SetUserFollowerCache(ctx context.Context, userID, followedUID int64, data *model.UserFollowerModel, duration time.Duration) error {
	if data == nil || userID == 0 {
		return nil
	}
	cacheKey := c.GetUserFollowerCacheKey(userID, followedUID)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// GetUserFollowerCache get from cache
func (c *userFollowerCache) GetUserFollowerCache(ctx context.Context, userID, followedUID int64) (data *model.UserFollowerModel, err error) {
	cacheKey := c.GetUserFollowerCacheKey(userID, followedUID)
	err = c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		log.WithContext(ctx).Warnf("get err from redis, err: %+v", err)
		return nil, err
	}
	return data, nil
}

// DelUserFollowerCache delete cache
func (c *userFollowerCache) DelUserFollowerCache(ctx context.Context, userID, followedUID int64) error {
	cacheKey := c.GetUserFollowerCacheKey(userID, followedUID)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}
