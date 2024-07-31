package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Filesharer is a Filesharer model.
type Filesharer struct {
	Hello string
}

// FilesharerRepo is a Greater repo.
type FilesharerRepo interface {
	Save(context.Context, *Filesharer) (*Filesharer, error)
	Update(context.Context, *Filesharer) (*Filesharer, error)
	FindByID(context.Context, int64) (*Filesharer, error)
	ListByHello(context.Context, string) ([]*Filesharer, error)
	ListAll(context.Context) ([]*Filesharer, error)
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
func (uc *FilesharerUsecase) CreateFilesharer(ctx context.Context, g *Filesharer) (*Filesharer, error) {
	uc.log.WithContext(ctx).Infof("CreateFilesharer: %v", g.Hello)
	return uc.repo.Save(ctx, g)
}
