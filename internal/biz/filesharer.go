package biz

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	pb "filesharer/api/file/v1"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pierrec/lz4"
	"google.golang.org/grpc"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	GetAllFiles(path string, s string) []FileInfo
}

type FileInfo struct {
	Path  string
	Size  int64
	Body  []byte
	IsDir bool
	Mode  fs.FileMode
	Fi    os.FileInfo
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
func (uc *FilesharerUsecase) DownloadByStream(stream grpc.ServerStreamingClient[pb.DownloadByAddrReply], path string) error {
	_ = os.MkdirAll("downloads", 0644)
	_, fileName := filepath.Split(path)
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
func (uc *FilesharerUsecase) DownloadDirByStream(stream grpc.ServerStreamingClient[pb.DownloadDirByAddrReply], path string) error {
	_ = os.MkdirAll("downloads", 0644)
	_, fileName := filepath.Split(path)
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
	//todo window 无法删除文件
	_ = os.RemoveAll(fileName)
	return nil

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
	var files = uc.repo.GetAllFiles(req.Path, "")
	for _, file := range files {
		hdr, err := tar.FileInfoHeader(file.Fi, "")
		if err != nil {
			return err
		}

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
