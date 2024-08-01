package service

import (
	"archive/tar"
	"context"
	"errors"
	pb "filesharer/api/file/v1"
	v1 "filesharer/api/file/v1"
	"filesharer/internal/biz"
	"filesharer/internal/data"
	"fmt"
	"github.com/pierrec/lz4"
	"github.com/todocoder/go-stream/stream"
	"io"
	"net/url"
	"strings"

	"os"
	"path/filepath"
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

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	//return nil
	buf := make([]byte, bufSize)
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
		return s.uc.DownloadDirByAddr(req, conn)
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
	_ = os.MkdirAll("downloads", 0644)
	_, fileName := filepath.Split(req.Path)
	fileName = filepath.Join("downloads", fileName) + ".tar"

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	//return nil
	buf := make([]byte, bufSize)
	//

	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
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

	rfile, err := os.Open(fileName)
	if err != nil {
		return errors.New("系统错误")
	}
	ext := filepath.Ext(fileName)

	dirName := strings.SplitN(fileName, ext, -1)[0]
	if len(dirName) == 0 {
		return errors.New("系统错误")

	}
	_ = os.MkdirAll(dirName, 0644)

	tr := tar.NewReader(rfile)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return errors.New("系统错误")
		}

		tarFile := filepath.Join(dirName, hdr.Name)
		mkdirDirString := "./" + filepath.Dir(tarFile)
		abs, _ := filepath.Abs(mkdirDirString)
		if !hdr.FileInfo().IsDir() {
			abs = filepath.Dir(abs)
		}

		err = os.MkdirAll(abs, 0777)
		if err != nil {
			panic(err)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		writeFileName := filepath.Join(abs, hdr.Name)
		os.MkdirAll(filepath.Dir(writeFileName), 0644)
		f, err := os.OpenFile(writeFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return errors.New("系统错误")
		}
		if _, err := io.CopyBuffer(f, tr, buf); err != nil {
			fmt.Printf("copy err %v\n", err)
		}
		f.Close()

	}
	rfile.Close()
	_ = os.RemoveAll(fileName)
	return nil

}

func (s *FileService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {

	return s.uc.ListNode(ctx, req)

}
