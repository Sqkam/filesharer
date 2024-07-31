package service

import (
	"context"
	pb "filesharer/api/file/v1"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/biz"
	"filesharer/internal/data"
	"fmt"
	"net/url"
	"sync"
)

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
func (s *FileService) DownloadByAddr(ctx context.Context, req *pb.DownloadByAddrRequest) (*pb.DownloadByAddrReply, error) {
	if Endpoint.Host == req.Addr {
		return s.uc.DownloadByAddr(ctx, req)
	}
	client, err := s.getClient(req.Addr)
	if err != nil {
		return nil, err
	}
	resp, err := client.DownloadByAddr(ctx, req)
	if err != nil {
		m.Delete(req.Addr)
	}
	return resp, err

}

func (s *FileService) DownloadDirByAddr(req *pb.DownloadDirByAddrRequest, conn pb.File_DownloadDirByAddrServer) error {
	if Endpoint.Host == req.Addr {
		for {
			err := conn.Send(&pb.DownloadDirByAddrReply{})
			if err != nil {
				return err
			}
		}
	}

	client, err := s.getClient(req.Addr)
	if err != nil {
		return err
	}
	stream, err := client.DownloadDirByAddr(context.Background(), req)
	if err != nil {
		m.Delete(req.Addr)
		return err
	}
	for {
		recv, err := stream.Recv()
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", recv)
	}
}

func (s *FileService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {

	return s.uc.ListNode(ctx, req)

}
