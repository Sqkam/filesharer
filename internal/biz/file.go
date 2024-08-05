package biz

import (
	"errors"
	pb "filesharer/api/file/v1"

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

	buf := make([]byte, bufSize)

	file, err := os.OpenFile(req.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	var n int
	for {
		n, err = file.Read(buf)
		if err != nil {
			break
		}
		err = conn.Send(&pb.DownloadByAddrReply{
			Data: buf[:n],
		})
		if err != nil {
			break
		}
	}
	return err
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

	return err
}
