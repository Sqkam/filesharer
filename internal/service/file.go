package service

import (
	"context"
	"errors"
	pb "filesharer/api/file/v1"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/biz"
	"filesharer/internal/data"
	"filesharer/third_party/snowflake"
	"fmt"
	"github.com/pierrec/lz4"
	"github.com/todocoder/go-stream/stream"
	"io"
	"net/url"
	"os"
	"path/filepath"
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
func (s *FileService) DownloadByAddr(req *pb.DownloadByAddrRequest, conn pb.File_DownloadByAddrServer) error {
	// 不会下载自己实例的文件
	node, err := s.uc.ListNode(context.Background(), &pb.ListNodeRequest{})
	if err != nil {
		return err
	}
	noMatch := stream.Of(node.Data...).NoneMatch(func(item *v1.ListNodeReplyItem) bool {
		return fmt.Sprintf("%s:%d", item.ServiceAddress, item.ServicePort) == req.Addr
	})
	if noMatch {
		return errors.New("非法addr")
	}

	if Endpoint.Host == req.Addr {
		return s.uc.DownloadByAddr(req, conn)
	}

	client, err := s.getClient(req.Addr)
	if err != nil {
		return err
	}
	stream, err := client.DownloadByAddr(context.Background(), req)
	if err != nil {
		m.Delete(req.Addr)
		return err
	}
	_ = os.MkdirAll("downloads", 0644)
	_, fileName := filepath.Split(req.Path)
	fileName = filepath.Join("downloads", fileName)
	_, err = os.Stat(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
		} else {
			return err
		}
	}

	if err == nil {
		fileName = filepath.Base(fileName) + "-" + snowflake.GenID() + filepath.Ext(fileName)
		fileName = filepath.Join("downloads", fileName)
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	//return nil
	buf := make([]byte, 8192)
	//

	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		block, err := lz4.UncompressBlock(recv.Data, buf)
		if err != nil {
			return err
		}

		_, err = file.Write(buf[:block])
		if err != nil {
			return err
		}
	}

}
func (s *FileService) DownloadDirByAddr(req *pb.DownloadDirByAddrRequest, conn pb.File_DownloadDirByAddrServer) error {
	file, err := os.OpenFile("/root/temp", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", names)
	return nil
	//如果是文件夹,还想解压了
}

func (s *FileService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {

	return s.uc.ListNode(ctx, req)

}
