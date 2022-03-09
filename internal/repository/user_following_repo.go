package repository

//go:generate mockgen -source=user_following_repo.go -destination=../../internal/mocks/user_following_repo_mock.go  -package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/go-eagle/eagle/pkg/redis"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/go-microservice/relation-service/internal/cache"
	"github.com/go-microservice/relation-service/internal/model"
)

var (
	_tableUserFollowingName   = (&model.UserFollowingModel{}).TableName()
	_getUserFollowingSQL      = "SELECT * FROM %s WHERE user_id = %d and followed_uid = %d"
	_batchGetUserFollowingSQL = "SELECT * FROM %s WHERE id IN (%s)"
)

var _ UserFollowingRepo = (*userFollowingRepo)(nil)

// UserFollowingRepo define a repo interface
type UserFollowingRepo interface {
	CreateUserFollowing(ctx context.Context, db *gorm.DB, data *model.UserFollowingModel) (id int64, err error)
	UpdateUserFollowingStatus(ctx context.Context, db *gorm.DB, userID, followedUID int64, status int) error
	GetUserFollowing(ctx context.Context, userID, followedUID int64) (ret *model.UserFollowingModel, err error)
}

type userFollowingRepo struct {
	db     *gorm.DB
	tracer trace.Tracer
	cache  cache.UserFollowingCache
}

// NewUserFollowing new a repository and return
func NewUserFollowing(db *gorm.DB, cache cache.UserFollowingCache) UserFollowingRepo {
	return &userFollowingRepo{
		db:     db,
		tracer: otel.Tracer("userFollowingRepo"),
		cache:  cache,
	}
}

// CreateUserFollowing create a item
func (r *userFollowingRepo) CreateUserFollowing(ctx context.Context, db *gorm.DB, data *model.UserFollowingModel) (id int64, err error) {
	err = db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return 0, errors.Wrap(err, "[repo] create UserFollowing err")
	}

	return data.ID, nil
}

// UpdateUserFollowing update item
func (r *userFollowingRepo) UpdateUserFollowingStatus(ctx context.Context, db *gorm.DB, userID, followedUID int64, status int) error {
	userFollow := model.UserFollowingModel{}
	err := db.Model(&userFollow).Where("user_id=? and followed_uid=?", userID, followedUID).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = r.cache.DelUserFollowingCache(ctx, userID, followedUID)
	return nil
}

// GetUserFollowing get a record
func (r *userFollowingRepo) GetUserFollowing(ctx context.Context, userID, followedUID int64) (ret *model.UserFollowingModel, err error) {
	// read cache
	item, err := r.cache.GetUserFollowingCache(ctx, userID, followedUID)
	if err != nil && !errors.Is(err, redis.ErrRedisNotFound) {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	data := new(model.UserFollowingModel)
	err = r.db.WithContext(ctx).Raw(fmt.Sprintf(_getUserFollowingSQL, _tableUserFollowingName, userID, followedUID)).Scan(&data).Error
	if err != nil {
		return
	}

	if data != nil && data.ID > 0 {
		err = r.cache.SetUserFollowingCache(ctx, userID, followedUID, data, 5*time.Minute)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
