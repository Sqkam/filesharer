package biz

import (
	"archive/tar"
	"bytes"

	"context"
	"errors"
	pb "filesharer/api/file/v1"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pierrec/lz4"
	"io"
	"os"
)

const bufSize = 8192 * 100 * 3

// Filesharer is a Filesharer model.
type Filesharer struct {
	Hello string
}

// FilesharerRepo is a Greater repo.
type FilesharerRepo interface {
	ListByAddr(ctx context.Context, req *pb.ListByAddrRequest) (*pb.ListByAddrReply, error)
	GetDetailByAddr(ctx context.Context, req *pb.GetDetailByAddrRequest) (*pb.GetDetailByAddrReply, error)
	DownloadByAddr(ctx context.Context, req *pb.DownloadByAddrRequest) (*pb.DownloadByAddrReply, error)

	ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error)
}

// FilesharerUsecase is a Filesharer usecase.
type FilesharerUsecase struct {
	repo FilesharerRepo
	log  *log.Helper
}

// NewFilesharerUsecase new a Filesharer usecase.
func NewFilesharerUsecase(repo FilesharerRepo, logger log.Logger) *FilesharerUsecase {
	return &FilesharerUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateFilesharer creates a Filesharer, and returns the new Filesharer.
func (uc *FilesharerUsecase) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {
	return uc.repo.ListNode(ctx, req)
}
func (uc *FilesharerUsecase) ListByAddr(ctx context.Context, req *pb.ListByAddrRequest) (*pb.ListByAddrReply, error) {
	return uc.repo.ListByAddr(ctx, req)
}
func (uc *FilesharerUsecase) GetDetailByAddr(ctx context.Context, req *pb.GetDetailByAddrRequest) (*pb.GetDetailByAddrReply, error) {
	return uc.repo.GetDetailByAddr(ctx, req)
}

func (uc *FilesharerUsecase) DownloadDirByAddr(req *pb.DownloadDirByAddrRequest, conn pb.File_DownloadDirByAddrServer) error {
	stat, err := os.Stat(req.Path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("不要乱搞")
	}
	var tarBuf bytes.Buffer

	tw := tar.NewWriter(&tarBuf)
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			panic(err)
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	readBuf := make([]byte, bufSize)
	lz4Buf := make([]byte, bufSize)
	ht := make([]int, 64<<10)
	for {
		n, err := tarBuf.Read(readBuf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		block, err := lz4.CompressBlock(readBuf[:n], lz4Buf, ht)

		err = conn.Send(&pb.DownloadDirByAddrReply{
			Data: lz4Buf[:block],
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *FilesharerUsecase) DownloadByAddr(req *pb.DownloadByAddrRequest, conn pb.File_DownloadByAddrServer) error {

	stat, err := os.Stat(req.Path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return errors.New("不要乱搞")
	}

	buf := make([]byte, bufSize)

	file, err := os.OpenFile(req.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	lz4Buf := make([]byte, len(buf))
	ht := make([]int, 64<<10)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		block, err := lz4.CompressBlock(buf[:n], lz4Buf, ht)
		err = conn.Send(&pb.DownloadByAddrReply{
			Data: lz4Buf[:block],
		})
		if err != nil {
			return err
		}
	}
}
