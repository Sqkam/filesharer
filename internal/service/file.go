package service

import (
	"context"
	pb "filesharer/api/file/v1"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/biz"
	"filesharer/internal/data"
	"net/url"
	"sync"
)

const bufSize = 8192 * 100 * 3

type FileService struct {
	pb.UnimplementedFileServer
	uc *biz.FilesharerUsecase
}

func NewFileService(uc *biz.FilesharerUsecase) *FileService {
	return &FileService{uc: uc}
}

var Endpoint *url.URL
var m = &sync.Map{}

func (s *FileService) getClient(addr string) (v1.FileClient, error) {
	value, ok := m.Load(addr)
	if ok {
		return value.(v1.FileClient), nil
	}
	client, err := data.NewFileClientWithAddr(addr)
	if err != nil {
		return nil, err
	}
	m.Store(addr, client)
	return client, nil
}
func (s *FileService) ListByAddr(ctx context.Context, req *pb.ListByAddrRequest) (*pb.ListByAddrReply, error) {
	if Endpoint.Host == req.Addr {
		return s.uc.ListByAddr(ctx, req)
	}
	client, err := s.getClient(req.Addr)
	if err != nil {
		return nil, err
	}
	resp, err := client.ListByAddr(ctx, req)
	if err != nil {
		m.Delete(req.Addr)
		return nil, err
	}

	return resp, err
}
func (s *FileService) GetDetailByAddr(ctx context.Context, req *pb.GetDetailByAddrRequest) (*pb.GetDetailByAddrReply, error) {
	if Endpoint.Host == req.Addr {
		return s.uc.GetDetailByAddr(ctx, req)
	}
	client, err := s.getClient(req.Addr)
	if err != nil {
		return nil, err
	}
	resp, err := client.GetDetailByAddr(ctx, req)
	if err != nil {
		m.Delete(req.Addr)
	}
	return resp, err
}
func (s *FileService) DownloadByAddr(req *pb.DownloadByAddrRequest, conn pb.File_DownloadByAddrServer) error {
	return s.uc.DownloadByAddr(req, conn)
}
func (s *FileService) DownloadDirByAddr(req *pb.DownloadDirByAddrRequest, conn pb.File_DownloadDirByAddrServer) error {
	return s.uc.DownloadDirByAddr(req, conn)
}

func (s *FileService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {

	return s.uc.ListNode(ctx, req)

}

func (s *FileService) DownloadByAddrHttp(ctx context.Context, req *pb.DownloadByAddrRequest) (*pb.DownloadByAddrReply, error) {
	client, err := s.getClient(req.Addr)
	if err != nil {
		return nil, err
	}

	stream, err := client.DownloadByAddr(context.Background(), req)
	if err != nil {
		m.Delete(req.Addr)
		return nil, err
	}
	err = s.uc.DownloadByStream(stream, req.Path)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s *FileService) DownloadDirByAddrHttp(ctx context.Context, req *pb.DownloadDirByAddrRequest) (*pb.DownloadDirByAddrReply, error) {
	client, err := s.getClient(req.Addr)
	if err != nil {
		return nil, err
	}
	stream, err := client.DownloadDirByAddr(ctx, req)
	if err != nil {
		m.Delete(req.Addr)
		return nil, err
	}
	err = s.uc.DownloadDirByStream(stream, req.Path)
	if err != nil {
		return nil, err
	}
	return nil, nil

}
