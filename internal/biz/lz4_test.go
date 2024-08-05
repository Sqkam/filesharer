package biz

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/Sqkam/gotools"
	"github.com/pierrec/lz4"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_lz4(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		path   string
		target string
	}{
		{
			path:   "./gin",
			target: "gin.tar.lz4",
			//path: "/root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			pr, pw, _ := os.Pipe()

			zw := lz4.NewWriter(pw)

			errCh := make(chan error)
			go func() {
				err = gotools.TarTo(tt.path, zw)

				zw.Close()
				pw.Close()
				errCh <- err
			}()

			file, err := os.OpenFile(tt.target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return
			}
			defer file.Close()
			_, err = io.Copy(file, pr)
			if err != nil {
				//return err
			}
			subErr := <-errCh
			if subErr != nil {
				//return subErr
			}
			//

		})
	}
}

func Test_tar(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		path   string
		target string
	}{
		{
			path:   "./gin",
			target: "gin.tar",
			//path: "/root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			pr, pw, _ := os.Pipe()

			zw := lz4.NewWriter(pw)
			zr := lz4.NewReader(pr)
			errCh := make(chan error)
			go func() {
				err = gotools.TarTo(tt.path, zw)

				zw.Close()
				pw.Close()
				errCh <- err
			}()

			file, err := os.OpenFile(tt.target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return
			}
			defer file.Close()
			_, err = io.Copy(file, zr)
			if err != nil {
				panic(err)
			}
			subErr := <-errCh
			if subErr != nil {
				panic(subErr)
			}
			//

		})
	}
}

func Test_tarSync(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		path string
	}{
		{
			path: "/root/temp/testgo",
			//path: "/root/temp",
			//path: "/root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarName := "asdf.tar"
			file, err := os.OpenFile(tarName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}

			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
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
				panic(err)
			}
			b := make([]byte, 8192*1000)

			_, _ = io.CopyBuffer(file, &buf, b)
			file.Close()

			rfile, err := os.Open(tarName)
			if err != nil {
				panic(err)
			}

			ext := filepath.Ext(tarName)
			dirName := tarName[:len(ext)]
			if len(dirName) == 0 {
				return

			}
			_ = os.MkdirAll(dirName, 0644)

			tr := tar.NewReader(rfile)

			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break // End of archive
				}
				if err != nil {
					return
				}

				tarFile := filepath.Join(dirName, hdr.Name)
				f, err := os.OpenFile(tarFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return
				}
				if _, err := io.CopyBuffer(f, tr, b); err != nil {
					fmt.Printf("%v\n", err)
				}
				f.Close()

			}
			rfile.Close()
			_ = os.RemoveAll(tarName)

		})
	}
}

func Test_tarDirect(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		path string
	}{
		{
			path: "/root/temp/testgo",
			//path: "/root/temp",
			//path: "/root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarName := "asdf.tar"
			file, err := os.OpenFile(tarName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}

			tw := tar.NewWriter(file)
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
				panic(err)
			}
			b := make([]byte, 8192*1000)

			file.Close()

			rfile, err := os.Open(tarName)
			if err != nil {
				panic(err)
			}

			ext := filepath.Ext(tarName)
			fmt.Printf("%v\n", len(ext))

			dirName := strings.SplitN(tarName, ext, -1)[0]
			if len(dirName) == 0 {
				return

			}
			_ = os.MkdirAll(dirName, 0644)

			tr := tar.NewReader(rfile)

			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break // End of archive
				}
				if err != nil {
					return
				}

				tarFile := filepath.Join(dirName, hdr.Name)
				f, err := os.OpenFile(tarFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return
				}
				if _, err := io.CopyBuffer(f, tr, b); err != nil {
					fmt.Printf("%v\n", err)
				}
				f.Close()

			}
			rfile.Close()
			_ = os.RemoveAll(tarName)

		})
	}
}
