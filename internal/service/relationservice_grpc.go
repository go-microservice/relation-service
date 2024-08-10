package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-eagle/eagle/pkg/errcode"

	pb "github.com/go-microservice/relation-service/api/relation/v1"
	"github.com/go-microservice/relation-service/internal/ecode"
	"github.com/go-microservice/relation-service/internal/model"
	repo "github.com/go-microservice/relation-service/internal/repository"
)

const (
	// FollowStatusNormal 关注状态-正常
	FollowStatusNormal int = 1 // 正常
	// FollowStatusDelete 关注状态-删除
	FollowStatusDelete = 0 // 删除
)

var (
	_ pb.RelationServiceServer = (*RelationServiceServer)(nil)
)

type RelationServiceServer struct {
	pb.UnimplementedRelationServiceServer

	followerRepo  repo.UserFollowerRepo
	followingRepo repo.UserFollowingRepo
}

func NewRelationServiceServer(followerRepo repo.UserFollowerRepo, followingRepo repo.UserFollowingRepo) *RelationServiceServer {
	return &RelationServiceServer{
		followerRepo:  followerRepo,
		followingRepo: followingRepo,
	}
}

// Follow user
func (s *RelationServiceServer) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowReply, error) {
	// if is follow self
	if isSelf(req.GetUserId(), req.GetFollowedUid()) {
		return nil, ecode.ErrInvalidArgument.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": errors.New("can not follow yourself"),
		})).Status(req).Err()
	}

	// check if has followed
	following, err := s.followingRepo.GetUserFollowing(ctx, req.UserId, req.FollowedUid)
	if err != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}
	// has follow
	if following != nil && following.Status == FollowStatusNormal {
		return &pb.FollowReply{}, nil
	}

	db := model.GetDB()
	tx := db.Begin()
	if tx.Error != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": tx.Error.Error(),
		})).Status(req).Err()
	}

	curTime := time.Now()
	// 添加到关注表
	_, err = s.followingRepo.CreateUserFollowing(ctx, tx, &model.UserFollowingModel{
		UserID:      req.UserId,
		FollowedUID: req.FollowedUid,
		Status:      FollowStatusNormal,
		CreatedAt:   curTime,
		UpdatedAt:   curTime,
	})
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}
	// 添加到粉丝表
	_, err = s.followerRepo.CreateUserFollower(ctx, tx, &model.UserFollowerModel{
		UserID:      req.FollowedUid,
		FollowerUID: req.UserId,
		Status:      FollowStatusNormal,
		CreatedAt:   curTime,
		UpdatedAt:   curTime,
	})
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	// 增加关注数

	// 增加粉丝数

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	return &pb.FollowReply{}, nil
}

// Unfollow
func (s *RelationServiceServer) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowReply, error) {
	// cannot unfollow self
	if isSelf(req.GetUserId(), req.GetFollowedUid()) {
		return nil, ecode.ErrInvalidArgument.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": errors.New("cannot unfollow self"),
		})).Status(req).Err()
	}

	// 已取关
	following, err := s.followingRepo.GetUserFollowingWithoutCache(ctx, req.UserId, req.FollowedUid)
	if err != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}
	if following != nil && following.Status == FollowStatusDelete {
		return &pb.UnfollowReply{}, nil
	}

	// 如果是已关注，执行取关逻辑
	db := model.GetDB()
	tx := db.Begin()
	if tx.Error != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": tx.Error.Error(),
		})).Status(req).Err()
	}
	// 删除关注
	err = s.followingRepo.UpdateUserFollowingStatus(ctx, tx, req.UserId, req.FollowedUid, FollowStatusDelete)
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	// 删除粉丝
	err = s.followerRepo.UpdateUserFollowerStatus(ctx, tx, req.FollowedUid, req.UserId, FollowStatusDelete)
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	// 减少关注数

	// 减少粉丝数

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	return &pb.UnfollowReply{}, nil
}

func isSelf(UId, otherUId int64) bool {
	return UId == otherUId
}

func (s *RelationServiceServer) BatchGetRelation(ctx context.Context, req *pb.BatchGetRelationRequest) (*pb.BatchGetRelationReply, error) {
	if req.GetUserId() == 0 || len(req.GetIds()) == 0 {
		return nil, ecode.ErrInvalidArgument.WithDetails().Status(req).Err()
	}

	ret, err := s.followingRepo.BatchGetUserFollowing(ctx, req.GetUserId(), req.GetIds())
	if err != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}

	retMap := make(map[int64]int64)
	for _, v := range ret {
		retMap[v.FollowedUID] = int64(v.Status)
	}

	return &pb.BatchGetRelationReply{
		Result: retMap,
	}, nil
}

func (s *RelationServiceServer) GetFollowingList(ctx context.Context, req *pb.FollowingListRequest) (*pb.FollowingListReply, error) {
	if req.GetLastId() == 0 {
		req.LastId = MaxID
	}
	userFollowList, err := s.followingRepo.GetFollowingUserList(ctx, req.UserId, req.LastId, int(req.Limit))
	if err != nil {
		return nil, err
	}

	var data []*pb.FollowingListReplyUserFollow
	for _, v := range userFollowList {
		item := pb.FollowingListReplyUserFollow{
			Id:          v.ID,
			FollowedUid: v.FollowedUID,
		}
		data = append(data, &item)
	}

	return &pb.FollowingListReply{
		Result: data,
	}, nil
}

func (s *RelationServiceServer) GetFollowerList(ctx context.Context, req *pb.FollowerListRequest) (*pb.FollowerListReply, error) {
	if req.GetLastId() == 0 {
		req.LastId = MaxID
	}
	userFollowList, err := s.followerRepo.GetFollowerUserList(ctx, req.UserId, req.LastId, int(req.Limit))
	if err != nil {
		return nil, err
	}

	var data []*pb.FollowerListReplyFollower
	for _, v := range userFollowList {
		item := pb.FollowerListReplyFollower{
			Id:          v.ID,
			FollowerUid: v.FollowerUID,
		}
		data = append(data, &item)
	}

	return &pb.FollowerListReply{
		Result: data,
	}, nil
}
