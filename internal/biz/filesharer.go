package biz

import (
	"archive/tar"
	"bytes"
	"io/fs"
	"path/filepath"
	"sync"

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
	var files = GetAllFiles(req.Path, "")
	for _, file := range files {
		//hdr := &tar.Header{
		//	Name:  file.Path,
		//	Mode: 0644,
		//	Size: file.Size,
		//}
		hdr, err := tar.FileInfoHeader(file.Fi, "")
		if err != nil {
			return err
		}
		//hdr.Name = strings.TrimPrefix(fileName, prefix)
		hdr.Name = file.Path

		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if _, err := tw.Write(file.Body); err != nil {
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
		if err != nil {
			return err
		}

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

type FileInfo struct {
	Path  string
	Size  int64
	Body  []byte
	IsDir bool
	Mode  fs.FileMode
	Fi    os.FileInfo
}
type FileItem struct {
	RelPaths string
	AbsPaths string
}

func GetAllFiles(path string, parent string) []FileInfo {
	if !filepath.IsAbs(path) {
		return nil
	}

	relPaths, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil
	}
	files := make([]*FileItem, len(relPaths))
	for i, v := range relPaths {
		abs, _ := filepath.Abs(v)
		files[i] = &FileItem{RelPaths: v, AbsPaths: abs}
	}

	resp := make([]FileInfo, 0)
	wg := &sync.WaitGroup{}
	ch := make(chan FileInfo, len(files))
	for _, v := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := v
			info, err := os.Stat(v.RelPaths)
			if err != nil {
				return
			}
			filePath := filepath.Join(parent, info.Name())
			if info.IsDir() {
				allFiles := GetAllFiles(v.RelPaths, filePath)
				for _, f := range allFiles {
					ch <- f
				}
			}

			f, err := os.Open(v.AbsPaths)
			if err != nil {
				return
			}
			defer f.Close()
			all := make([]byte, 0)
			if !info.IsDir() {
				all, err = io.ReadAll(f)
				if err != nil {
					return
				}
			}
			info.Mode()
			ch <- FileInfo{
				Path:  filePath,
				IsDir: info.IsDir(),
				Size:  info.Size(),
				Body:  all,
				Mode:  info.Mode(),
				Fi:    info,
			}
		}()
	}

	chDone := make(chan struct{})
	go func() {
		for v := range ch {
			resp = append(resp, v)
		}

		chDone <- struct{}{}
	}()
	wg.Wait()
	close(ch)
	<-chDone
	return resp
}
