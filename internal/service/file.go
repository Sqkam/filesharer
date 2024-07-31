package service

import (
	"context"

	pb "filesharer/api/file/v1"
)

type FileService struct {
	pb.UnimplementedFileServer
}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) ListByHostname(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	return &pb.ListFileReply{}, nil
}
func (s *FileService) GetDetailByHostname(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	return &pb.ListFileReply{}, nil
}
func (s *FileService) DownloadByHostname(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	return &pb.ListFileReply{}, nil
}
func (s *FileService) DownloadDirByHostname(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	return &pb.ListFileReply{}, nil
}
func (s *FileService) ListNode(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	return &pb.ListFileReply{}, nil
}
