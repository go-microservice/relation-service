package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-microservice/relation-service/internal/model"

	"github.com/go-microservice/relation-service/internal/ecode"

	"github.com/go-eagle/eagle/pkg/errcode"

	"github.com/google/wire"

	pb "github.com/go-microservice/relation-service/api/relation/v1"
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

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewRelationServiceServer)

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

func (s *RelationServiceServer) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowReply, error) {
	// check if is self
	if req.GetUserId() == req.GetFollowedUid() {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
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
	if following != nil && following.Status == FollowStatusNormal {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": errors.New("has following"),
		})).Status(req).Err()
	}

	// follow logic
	_, err = s.followingRepo.CreateUserFollowing(ctx, &model.UserFollowingModel{
		UserID:      req.UserId,
		FollowedUID: req.FollowedUid,
		Status:      FollowStatusNormal,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	})
	if err != nil {
		return nil, ecode.ErrInternalError.WithDetails(errcode.NewDetails(map[string]interface{}{
			"msg": err.Error(),
		})).Status(req).Err()
	}
	return &pb.FollowReply{}, nil
}
func (s *RelationServiceServer) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowReply, error) {
	return &pb.UnfollowReply{}, nil
}
func (s *RelationServiceServer) BatchGetRelation(ctx context.Context, req *pb.BatchGetRelationRequest) (*pb.BatchGetRelationReply, error) {
	return &pb.BatchGetRelationReply{}, nil
}
func (s *RelationServiceServer) GetFollowingList(ctx context.Context, req *pb.FollowingListRequest) (*pb.FollowingListReply, error) {
	return &pb.FollowingListReply{}, nil
}
func (s *RelationServiceServer) GetFollowerList(ctx context.Context, req *pb.FollowerListRequest) (*pb.FollowerListRequest, error) {
	return &pb.FollowerListRequest{}, nil
}
