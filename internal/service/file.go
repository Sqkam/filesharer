package service

import (
	"context"
	"filesharer/internal/biz"

	pb "filesharer/api/file/v1"
)

type FileService struct {
	pb.UnimplementedFileServer
	uc *biz.FilesharerUsecase
}

func NewFileService(uc *biz.FilesharerUsecase) *FileService {
	return &FileService{uc: uc}
}

func (s *FileService) ListByIp(ctx context.Context, req *pb.ListByIpRequest) (*pb.ListByIpReply, error) {
	return &pb.ListByIpReply{}, nil
}
func (s *FileService) GetDetailByIp(ctx context.Context, req *pb.GetDetailByIpRequest) (*pb.GetDetailByIpReply, error) {
	return &pb.GetDetailByIpReply{}, nil
}
func (s *FileService) DownloadByIp(ctx context.Context, req *pb.DownloadByIpRequest) (*pb.DownloadByIpReply, error) {
	return &pb.DownloadByIpReply{}, nil
}
func (s *FileService) DownloadDirByIp(ctx context.Context, req *pb.DownloadDirByIpRequest) (*pb.DownloadDirByIpReply, error) {
	return &pb.DownloadDirByIpReply{}, nil
}
func (s *FileService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {
	return &pb.ListNodeReply{}, nil
}
