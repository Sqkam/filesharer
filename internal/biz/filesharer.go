package biz

import (
	"context"
	pb "filesharer/api/file/v1"

	"github.com/go-kratos/kratos/v2/log"
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
	DownloadDirByAddr(ctx context.Context, req *pb.DownloadDirByAddrRequest) (*pb.DownloadDirByAddrReply, error)
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
func (uc *FilesharerUsecase) DownloadByAddr(ctx context.Context, req *pb.DownloadByAddrRequest) (*pb.DownloadByAddrReply, error) {
	return uc.repo.DownloadByAddr(ctx, req)
}
func (uc *FilesharerUsecase) DownloadDirByAddr(ctx context.Context, req *pb.DownloadDirByAddrRequest) (*pb.DownloadDirByAddrReply, error) {
	return uc.repo.DownloadDirByAddr(ctx, req)
}
