package service

import "github.com/google/wire"

const (
	// MaxID 最大id
	MaxID = 0xffffffffffff
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewRelationServiceServer)
