package biz

import (
	"context"
	pb "filesharer/api/file/v1"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	"os"
)

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
	for {
		err := conn.Send(&pb.DownloadDirByAddrReply{})
		if err != nil {
			return err
		}
	}
}

func (uc *FilesharerUsecase) DownloadByAddr(req *pb.DownloadByAddrRequest, conn pb.File_DownloadByAddrServer) error {
	b := make([]byte, 8192)
	file, err := os.OpenFile(req.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	//file.Readdirnames()
	for {
		n, err := file.Read(b)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = conn.Send(&pb.DownloadByAddrReply{
			Data: b[:n],
		})
		if err != nil {
			return err
		}
	}
}
