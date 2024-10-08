// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-eagle/eagle/pkg/app"
	"github.com/go-eagle/eagle/pkg/client/consulclient"
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/go-eagle/eagle/pkg/redis"
	"github.com/go-eagle/eagle/pkg/registry"
	"github.com/go-eagle/eagle/pkg/registry/consul"
	"github.com/go-eagle/eagle/pkg/transport/grpc"
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

func InitApp(cfg *app.Config, config *app.ServerConfig) (*app.App, func(), error) {
	db := model.GetDB()
	client, cleanup, err := redis.Init()
	if err != nil {
		return nil, nil, err
	}
	userFollowerCache := cache.NewUserFollowerCache(client)
	userFollowerRepo := repository.NewUserFollower(db, userFollowerCache)
	userFollowingCache := cache.NewUserFollowingCache(client)
	userFollowingRepo := repository.NewUserFollowing(db, userFollowingCache)
	relationServiceServer := service.NewRelationServiceServer(userFollowerRepo, userFollowingRepo)
	grpcServer := server.NewGRPCServer(config, relationServiceServer)
	appApp := newApp(cfg, grpcServer)
	return appApp, func() {
		cleanup()
	}, nil
}

// wire.go:

func newApp(cfg *app.Config, gs *grpc.Server) *app.App {
	return app.New(app.WithName(cfg.Name), app.WithVersion(cfg.Version), app.WithLogger(log.GetLogger()), app.WithServer(server.NewHTTPServer(&cfg.HTTP), gs), app.WithRegistry(getConsulRegistry()),
	)
}

// create a consul register
func getConsulRegistry() registry.Registry {
	client, err := consulclient.New()
	if err != nil {
		panic(err)
	}
	return consul.New(client)
}
