package cache

//go:generate mockgen -source=internal/cache/user_following_cache.go -destination=internal/mock/user_following_cache_mock.go  -package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-eagle/eagle/pkg/cache"
	"github.com/go-eagle/eagle/pkg/encoding"
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/go-redis/redis/v8"

	"github.com/go-microservice/relation-service/internal/model"
)

const (
	// PrefixUserFollowingCacheKey cache prefix
	PrefixUserFollowingCacheKey = "user:following:%d_%d"
)

// UserFollowing define cache interface
type UserFollowingCache interface {
	SetUserFollowingCache(ctx context.Context, userID, followedUID int64, data *model.UserFollowingModel, duration time.Duration) error
	GetUserFollowingCache(ctx context.Context, userID, followedUID int64) (data *model.UserFollowingModel, err error)
	DelUserFollowingCache(ctx context.Context, userID, followedUID int64) error
}

// userFollowingCache define cache struct
type userFollowingCache struct {
	cache cache.Cache
}

// NewUserFollowingCache new a cache
func NewUserFollowingCache(rdb *redis.Client) UserFollowingCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""
	return &userFollowingCache{
		cache: cache.NewRedisCache(rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UserFollowingModel{}
		}),
	}
}

// GetUserFollowingCacheKey get cache key
func (c *userFollowingCache) GetUserFollowingCacheKey(userID, followedUID int64) string {
	return fmt.Sprintf(PrefixUserFollowingCacheKey, userID, followedUID)
}

// SetUserFollowingCache write to cache
func (c *userFollowingCache) SetUserFollowingCache(ctx context.Context, userID, followedUID int64, data *model.UserFollowingModel, duration time.Duration) error {
	if data == nil || userID == 0 {
		return nil
	}
	cacheKey := c.GetUserFollowingCacheKey(userID, followedUID)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// GetUserFollowingCache get from cache
func (c *userFollowingCache) GetUserFollowingCache(ctx context.Context, userID, followedUID int64) (data *model.UserFollowingModel, err error) {
	cacheKey := c.GetUserFollowingCacheKey(userID, followedUID)
	err = c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		log.WithContext(ctx).Warnf("get err from redis, err: %+v", err)
		return nil, err
	}
	return data, nil
}

// DelUserFollowingCache delete cache
func (c *userFollowingCache) DelUserFollowingCache(ctx context.Context, userID, followedUID int64) error {
	cacheKey := c.GetUserFollowingCacheKey(userID, followedUID)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// DelUserFollowingCache set empty cache
func (c *userFollowingCache) SetCacheWithNotFound(ctx context.Context, userID, followedUID int64) error {
	cacheKey := c.GetUserFollowingCacheKey(userID, followedUID)
	err := c.cache.SetCacheWithNotFound(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}
