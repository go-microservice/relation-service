package ecode

import (
	"github.com/go-eagle/eagle/pkg/errcode"
	"google.golang.org/grpc/codes"
)

//nolint: golint
var (
	// common errors
	ErrInvalidArgument = errcode.New(codes.InvalidArgument, "Invalid argument")
	ErrInternalError   = errcode.New(codes.Internal, "Internal error")
	ErrAccessDenied    = errcode.New(codes.PermissionDenied, "Access denied")
	ErrNotFound        = errcode.New(codes.NotFound, "Not found")

	// relation grpc errors
	ErrUserIsExist = errcode.New(20100, "The user already exists.")
)
