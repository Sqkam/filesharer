package data

import (
	"context"

	"filesharer/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type FilesharerRepo struct {
	data *Data
	log  *log.Helper
}

// NewFilesharerRepo .
func NewFilesharerRepo(data *Data, logger log.Logger) biz.FilesharerRepo {
	return &FilesharerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *FilesharerRepo) Save(ctx context.Context, g *biz.Filesharer) (*biz.Filesharer, error) {
	return g, nil
}

func (r *FilesharerRepo) Update(ctx context.Context, g *biz.Filesharer) (*biz.Filesharer, error) {
	return g, nil
}

func (r *FilesharerRepo) FindByID(context.Context, int64) (*biz.Filesharer, error) {
	return nil, nil
}

func (r *FilesharerRepo) ListByHello(context.Context, string) ([]*biz.Filesharer, error) {
	return nil, nil
}

func (r *FilesharerRepo) ListAll(context.Context) ([]*biz.Filesharer, error) {
	return nil, nil
}
