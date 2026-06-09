//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"eino-stock/internal/biz"
	"eino-stock/internal/conf"
	"eino-stock/internal/data"
	"eino-stock/internal/infrastructure"
	"eino-stock/internal/server"
	"eino-stock/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.DataSource, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, infrastructure.ProviderSet, newApp))
}
