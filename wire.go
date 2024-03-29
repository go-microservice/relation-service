// +build wireinject

package main

import (
	eagle "github.com/go-eagle/eagle/pkg/app"
	"github.com/go-microservice/relation-service/internal/cache"
	"github.com/go-microservice/relation-service/internal/repository"
	"github.com/go-microservice/relation-service/internal/server"
	"github.com/go-microservice/relation-service/internal/service"
	"github.com/google/wire"
)

func InitApp(cfg *eagle.Config, config *eagle.ServerConfig) (*eagle.App, error) {
	wire.Build(server.ProviderSet, service.ProviderSet, repository.ProviderSet, cache.ProviderSet, newApp)
	return &eagle.App{}, nil
}
