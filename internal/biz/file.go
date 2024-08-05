package biz

import (
	"errors"
	pb "filesharer/api/file/v1"
	"fmt"
	"github.com/pierrec/lz4"
	"io"

	"google.golang.org/grpc"
	"os"
	"path/filepath"
)

func (uc *FilesharerUsecase) DownloadByAddr(req *pb.DownloadByAddrRequest, conn pb.File_DownloadByAddrServer) error {

	stat, err := os.Stat(req.Path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return errors.New("不要乱搞")
	}

	file, err := os.OpenFile(req.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	pr, pw, _ := os.Pipe()
	zw := lz4.NewWriter(pw)

	errCh := make(chan error)
	go func() {
		var err error
		_, err = io.Copy(zw, file)

		zw.Close()
		pw.Close()
		errCh <- err

	}()

	writer := NewGrpcWriter(conn)
	_, err = io.Copy(writer, pr)
	if err != nil {
		return err
	}
	err = <-errCh
	if err != nil {
		return err
	}
	return nil

}

func (uc *FilesharerUsecase) DownloadByStream(stream grpc.ServerStreamingClient[pb.DownloadByAddrReply], path string) error {
	_ = os.MkdirAll(SaveDir, 0644)
	_, fileName := filepath.Split(path)
	fileName = filepath.Join(SaveDir, fileName)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	pr, pw := io.Pipe()

	zr := lz4.NewReader(pr)
	errCh := make(chan error)

	var readCount int64
	var writeCount int64
	go func() {
		reader := NewGrpcReader(stream)
		var err error

		readCount, err = io.Copy(pw, reader)

		_ = pw.Close()
		errCh <- err
	}()

	writeCount, err = io.Copy(file, zr)
	if err != nil {
		return err
	}
	err = <-errCh
	if err != nil {
		return err
	}
	fmt.Printf("file压缩率: %v\n", float64(readCount)/float64(writeCount))

	return nil
}
