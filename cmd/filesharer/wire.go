//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"filesharer/internal/biz"
	"filesharer/internal/conf"
	"filesharer/internal/data"
	"filesharer/internal/server"
	"filesharer/internal/service"
	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		InitRegistry,
		newApp))
}
func InitRegistry(bc *conf.Bootstrap) registry.Registrar {
	cfg := bc.Consul
	// new consul client
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = cfg.Addr
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client, consul.WithHealthCheck(true))

	return reg
}
