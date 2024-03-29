// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/go-eagle/eagle/pkg/app"
	"github.com/go-eagle/eagle/pkg/redis"
	"github.com/go-microservice/relation-service/internal/cache"
	"github.com/go-microservice/relation-service/internal/model"
	"github.com/go-microservice/relation-service/internal/repository"
	"github.com/go-microservice/relation-service/internal/server"
	"github.com/go-microservice/relation-service/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

func InitApp(cfg *app.Config, config *app.ServerConfig) (*app.App, error) {
	db := model.GetDB()
	client := redis.Init()
	userFollowerCache := cache.NewUserFollowerCache(client)
	userFollowerRepo := repository.NewUserFollower(db, userFollowerCache)
	userFollowingCache := cache.NewUserFollowingCache(client)
	userFollowingRepo := repository.NewUserFollowing(db, userFollowingCache)
	relationServiceServer := service.NewRelationServiceServer(userFollowerRepo, userFollowingRepo)
	grpcServer := server.NewGRPCServer(config, relationServiceServer)
	appApp := newApp(cfg, grpcServer)
	return appApp, nil
}
