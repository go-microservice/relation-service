package repository

import (
	"github.com/go-microservice/relation-service/internal/model"
	"github.com/google/wire"
)

// ProviderSet is repo providers.
var ProviderSet = wire.NewSet(model.Init())
