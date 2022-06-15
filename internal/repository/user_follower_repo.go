package repository

//go:generate mockgen -source=user_follower_repo.go -destination=../../internal/mocks/user_follower_repo_mock.go  -package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/go-microservice/relation-service/internal/cache"
	"github.com/go-microservice/relation-service/internal/model"
)

var (
	_tableUserFollowerName   = (&model.UserFollowerModel{}).TableName()
	_insertUserFollowerSQL   = "INSERT INTO %s SET user_id = ?, follower_uid =?, created_at = ?, status = ? on duplicate key update status = ?, updated_at = ?"
	_getUserFollowerSQL      = "SELECT * FROM %s WHERE id = ?"
	_batchGetUserFollowerSQL = "SELECT * FROM %s WHERE id IN (%s)"
)

var _ UserFollowerRepo = (*userFollowerRepo)(nil)

// UserFollowerRepo define a repo interface
type UserFollowerRepo interface {
	CreateUserFollower(ctx context.Context, db *gorm.DB, data *model.UserFollowerModel) (id int64, err error)
	UpdateUserFollowerStatus(ctx context.Context, db *gorm.DB, userID, followerUID int64, status int) error
	GetUserFollower(ctx context.Context, userID, followedUID int64) (ret *model.UserFollowerModel, err error)
	// 获取粉丝用户列表
	GetFollowerUserList(ctx context.Context, userID, lastID int64, limit int) ([]*model.UserFollowerModel, error)
}

type userFollowerRepo struct {
	db     *gorm.DB
	tracer trace.Tracer
	cache  cache.UserFollowerCache
}

// NewUserFollower new a repository and return
func NewUserFollower(db *gorm.DB, cache cache.UserFollowerCache) UserFollowerRepo {
	return &userFollowerRepo{
		db:     db,
		tracer: otel.Tracer("userFollowerRepo"),
		cache:  cache,
	}
}

// CreateUserFollower create a item
func (r *userFollowerRepo) CreateUserFollower(ctx context.Context, db *gorm.DB, data *model.UserFollowerModel) (id int64, err error) {
	_sql := fmt.Sprintf(_insertUserFollowerSQL, _tableUserFollowerName)
	err = db.WithContext(ctx).Exec(_sql,
		data.UserID, data.FollowerUID,
		data.CreatedAt, data.Status,
		data.Status, data.UpdatedAt,
	).Error
	if err != nil {
		return 0, errors.Wrap(err, "[repo] create UserFollower err")
	}

	return data.ID, nil
}

// UpdateUserFollower update item
func (r *userFollowerRepo) UpdateUserFollowerStatus(ctx context.Context, db *gorm.DB, userID, followerUID int64, status int) error {
	userFans := model.UserFollowerModel{}
	err := db.Model(&userFans).Where("user_id=? and follower_uid=?", userID, followerUID).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
	if err != nil {
		return err
	}
	// delete cache
	_ = r.cache.DelUserFollowerCache(ctx, userID, followerUID)
	return nil
}

// GetUserFollower get a record
func (r *userFollowerRepo) GetUserFollower(ctx context.Context, userID, followedUID int64) (ret *model.UserFollowerModel, err error) {
	// read cache
	item, err := r.cache.GetUserFollowerCache(ctx, userID, followedUID)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	data := new(model.UserFollowerModel)
	err = r.db.WithContext(ctx).Raw(fmt.Sprintf(_getUserFollowerSQL, _tableUserFollowerName), userID, followedUID).Scan(&data).Error
	if err != nil {
		return
	}

	if data.ID > 0 {
		err = r.cache.SetUserFollowerCache(ctx, userID, followedUID, data, 5*time.Minute)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// GetFollowingUserList 获取粉丝用户列表
func (r *userFollowerRepo) GetFollowerUserList(ctx context.Context, userID, lastID int64, limit int) ([]*model.UserFollowerModel, error) {
	userFollowerList := make([]*model.UserFollowerModel, 0)
	result := r.db.WithContext(ctx).Where("user_id=? AND id<=? and status=1", userID, lastID).
		Order("id desc").
		Limit(limit).Find(&userFollowerList)

	if err := result.Error; err != nil {
		return nil, errors.Wrapf(err, "get user follower list err")
	}

	return userFollowerList, nil
}
