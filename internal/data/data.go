package data

import (
	"context"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/conf"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewFilesharerRepo, NewFileClient)

// Data .
type Data struct {
	fileSharerClient v1.FileClient
}

func NewFileClient(bc *conf.Bootstrap) v1.FileClient {
	cfg := api.DefaultConfig()
	cfg.Address = bc.Consul.Addr
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	dis := consul.New(client)

	endpoint := "discovery:///review-service"
	conn, err := grpc.DialInsecure(context.Background(),

		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(dis))
	if err != nil {
		panic(err)
	}

	return v1.NewFileClient(conn)
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
