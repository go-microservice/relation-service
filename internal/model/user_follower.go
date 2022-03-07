package model

import "time"

// UserFollowerModel 粉丝表
type UserFollowerModel struct {
	ID          int64     `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"-"`
	UserID      int64     `gorm:"column:user_id" json:"user_id"`
	FollowerUID int64     `gorm:"column:follower_uid" json:"follower_uid"`
	Status      int       `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"-"`
}

// TableName sets the insert table name for this struct type
func (u *UserFollowerModel) TableName() string {
	return "user_follower"
}
