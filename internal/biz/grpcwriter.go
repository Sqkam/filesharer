package biz

import (
	pb "filesharer/api/file/v1"
	"google.golang.org/grpc"
)

type GrpcWriter struct {
	grpc.ServerStreamingServer[pb.DownloadByAddrReply]
}

func NewGrpcWriter(serverStreamingServer grpc.ServerStreamingServer[pb.DownloadByAddrReply]) *GrpcWriter {
	return &GrpcWriter{ServerStreamingServer: serverStreamingServer}
}

func (s *GrpcWriter) Write(p []byte) (n int, err error) {
	err = s.Send(&pb.DownloadByAddrReply{
		Data: p,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
