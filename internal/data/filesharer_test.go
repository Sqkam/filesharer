package data

import (
	"fmt"
	"testing"
	"time"
)

func Test_getAllFiles(t *testing.T) {
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
			now := time.Now()
			files := getAllFiles(tt.path)
			fmt.Printf("%v\n", len(files))
			fmt.Printf("spend:%v\n", time.Since(now).Milliseconds())
			//now2:=time.Now()
			//files2 := getAllFilesByWalk(tt.path)
			//fmt.Printf("%v\n", len(files2))
			//fmt.Printf("spend:%v\n", time.Since(now2).Milliseconds())
		})
	}
}
