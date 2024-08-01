package biz

import (
	pb "filesharer/api/file/v1"
	"google.golang.org/grpc"
	"io"
)

type GrpcReader struct {
	grpc.ServerStreamingClient[pb.DownloadByAddrReply]
}

func (g *GrpcReader) Read(p []byte) (n int, err error) {
	recv, err := g.ServerStreamingClient.Recv()
	if err != nil {
		return 0, err
	}
	n = copy(p, recv.Data)
	return n, nil

}

func NewGrpcReader(client grpc.ServerStreamingClient[pb.DownloadByAddrReply]) io.Reader {
	return &GrpcReader{ServerStreamingClient: client}

}
