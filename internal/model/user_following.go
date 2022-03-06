package model

import "time"

// UserFollowingModel 关注表
type UserFollowingModel struct {
	ID          uint64    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"-"`
	UserID      uint64    `gorm:"column:user_id" json:"user_id"`
	FollowedUID uint64    `gorm:"column:followed_uid" json:"followed_uid"`
	Status      int       `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"-"`
}

// TableName sets the insert table name for this struct type
func (u *UserFollowingModel) TableName() string {
	return "user_following"
}
