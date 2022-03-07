package repository

//go:generate mockgen -source=user_follower_repo.go -destination=../../internal/mocks/user_follower_repo_mock.go  -package mocks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/go-microservice/relation-service/internal/cache"
	"github.com/go-microservice/relation-service/internal/model"
)

var (
	_tableUserFollowerName   = (&model.UserFollowerModel{}).TableName()
	_getUserFollowerSQL      = "SELECT * FROM %s WHERE id = ?"
	_batchGetUserFollowerSQL = "SELECT * FROM %s WHERE id IN (%s)"
)

var _ UserFollowerRepo = (*userFollowerRepo)(nil)

// UserFollowerRepo define a repo interface
type UserFollowerRepo interface {
	CreateUserFollower(ctx context.Context, data *model.UserFollowerModel) (id int64, err error)
	UpdateUserFollower(ctx context.Context, id int64, data *model.UserFollowerModel) error
	GetUserFollower(ctx context.Context, id int64) (ret *model.UserFollowerModel, err error)
	BatchGetUserFollower(ctx context.Context, ids []int64) (ret []*model.UserFollowerModel, err error)
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
func (r *userFollowerRepo) CreateUserFollower(ctx context.Context, data *model.UserFollowerModel) (id int64, err error) {
	err = r.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return 0, errors.Wrap(err, "[repo] create UserFollower err")
	}

	return data.ID, nil
}

// UpdateUserFollower update item
func (r *userFollowerRepo) UpdateUserFollower(ctx context.Context, id int64, data *model.UserFollowerModel) error {
	item, err := r.GetUserFollower(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "[repo] update UserFollower err: %v", err)
	}
	err = r.db.Model(&item).Updates(data).Error
	if err != nil {
		return err
	}
	// delete cache
	_ = r.cache.DelUserFollowerCache(ctx, id)
	return nil
}

// GetUserFollower get a record
func (r *userFollowerRepo) GetUserFollower(ctx context.Context, id int64) (ret *model.UserFollowerModel, err error) {
	// read cache
	item, err := r.cache.GetUserFollowerCache(ctx, id)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	data := new(model.UserFollowerModel)
	err = r.db.WithContext(ctx).Raw(fmt.Sprintf(_getUserFollowerSQL, _tableUserFollowerName), id).Scan(&data).Error
	if err != nil {
		return
	}

	if data.ID > 0 {
		err = r.cache.SetUserFollowerCache(ctx, id, data, 5*time.Minute)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// BatchGetUserFollower batch get items
func (r *userFollowerRepo) BatchGetUserFollower(ctx context.Context, ids []int64) (ret []*model.UserFollowerModel, err error) {
	idsStr := cast.ToStringSlice(ids)
	itemMap, err := r.cache.MultiGetUserFollowerCache(ctx, ids)
	if err != nil {
		return nil, err
	}
	var missedID []int64
	for _, v := range ids {
		item, ok := itemMap[cast.ToString(v)]
		if !ok {
			missedID = append(missedID, v)
			continue
		}
		ret = append(ret, item)
	}
	// get missed data
	if len(missedID) > 0 {
		var missedData []*model.UserFollowerModel
		_sql := fmt.Sprintf(_batchGetUserFollowerSQL, _tableUserFollowerName, strings.Join(idsStr, ","))
		err = r.db.WithContext(ctx).Raw(_sql).Scan(&missedData).Error
		if err != nil {
			// you can degrade to ignore error
			return nil, err
		}
		if len(missedData) > 0 {
			ret = append(ret, missedData...)
			err = r.cache.MultiSetUserFollowerCache(ctx, missedData, 5*time.Minute)
			if err != nil {
				// you can degrade to ignore error
				return nil, err
			}
		}
	}
	return ret, nil
}
