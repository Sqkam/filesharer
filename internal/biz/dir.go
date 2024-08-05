package biz

import (
	"archive/tar"
	"errors"
	pb "filesharer/api/file/v1"
	"fmt"
	"github.com/Sqkam/gotools"
	"google.golang.org/grpc"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (uc *FilesharerUsecase) DownloadDirByStream(stream grpc.ServerStreamingClient[pb.DownloadDirByAddrReply], path string) error {
	_ = os.MkdirAll(SaveDir, 0644)
	_, fileName := filepath.Split(path)
	fileName = filepath.Join(SaveDir, fileName) + ".tar"

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
			break
		}
		_, err = file.Write(recv.Data)
		if err != nil {
			break
		}
	}

	return nil
	rfile, err := os.Open(fileName)
	if err != nil {
		return errors.New("系统错误")
	}
	defer rfile.Close()
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
			break
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

func (uc *FilesharerUsecase) DownloadDirByAddr(req *pb.DownloadDirByAddrRequest, conn pb.File_DownloadDirByAddrServer) error {
	stat, err := os.Stat(req.Path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("不要乱搞")
	}

	pr, pw, _ := os.Pipe()
	go func() {
		defer pw.Close()
		err = gotools.TarTo(req.Path, pw, false, func(s string, i ...interface{}) {})
		if err != nil {

			return
		}

	}()

	readBuf := make([]byte, bufSize)
	var n int
	for {
		n, err = pr.Read(readBuf)
		if err != nil {
			break
		}
		err = conn.Send(&pb.DownloadDirByAddrReply{
			Data: readBuf[:n],
		})
		if err != nil {
			break
		}
	}

	return err
}
