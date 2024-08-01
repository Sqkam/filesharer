package biz

import (
	pb "filesharer/api/file/v1"
	"google.golang.org/grpc"
	"io"
)

type GrpcDirWriter struct {
	ServerStreamingServer grpc.ServerStreamingServer[pb.DownloadDirByAddrReply]
}

func NewGrpcDirWriter(serverStreamingServer grpc.ServerStreamingServer[pb.DownloadDirByAddrReply]) io.Writer {
	return &GrpcDirWriter{ServerStreamingServer: serverStreamingServer}
}

func (s *GrpcDirWriter) Write(p []byte) (n int, err error) {
	err = s.ServerStreamingServer.Send(&pb.DownloadDirByAddrReply{
		Data: p,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
