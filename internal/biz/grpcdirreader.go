package biz

import (
	pb "filesharer/api/file/v1"
	"google.golang.org/grpc"
	"io"
)

type GrpcDirReader struct {
	grpc.ServerStreamingClient[pb.DownloadDirByAddrReply]
}

func (g *GrpcDirReader) Read(p []byte) (n int, err error) {
	recv, err := g.ServerStreamingClient.Recv()
	if err != nil {
		return 0, err
	}
	n = copy(p, recv.Data)
	return n, nil

}

func NewGrpcDirReader(client grpc.ServerStreamingClient[pb.DownloadDirByAddrReply]) io.Reader {
	return &GrpcDirReader{ServerStreamingClient: client}

}
