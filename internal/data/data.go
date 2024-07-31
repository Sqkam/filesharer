package data

import (
	"context"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/conf"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewFilesharerRepo, NewFileClient, NewGetConsulInfoClient)

type GetConsulInfoClient struct {
	ip string
}

func NewGetConsulInfoClient(bc *conf.Bootstrap) *GetConsulInfoClient {
	return &GetConsulInfoClient{ip: bc.Consul.Addr}
}

// Data .
type Data struct {
	fileSharerClient    v1.FileClient
	getConsulInfoClient *GetConsulInfoClient
}

func NewFileClient(bc *conf.Bootstrap) v1.FileClient {
	cfg := api.DefaultConfig()
	cfg.Address = bc.Consul.Addr
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	dis := consul.New(client)

	endpoint := "discovery:///filesharer"
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(dis))
	if err != nil {
		panic(err)
	}
	return v1.NewFileClient(conn)
}

func NewFileClientWithAddr(addr string) (v1.FileClient, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialInsecure(ctx,
		grpc.WithHealthCheck(true),
		grpc.WithEndpoint(addr),
	)
	if err != nil {
		return nil, err
	}
	return v1.NewFileClient(conn), nil
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, fileClient v1.FileClient, infoClient *GetConsulInfoClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		fileSharerClient:    fileClient,
		getConsulInfoClient: infoClient,
	}, cleanup, nil
}
