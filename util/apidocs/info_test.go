// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
)

func TestNewInfo(t *testing.T) {
	type args struct {
		fd *descriptor.FileDescriptor
	}
	tests := []struct {
		name     string
		args     args
		want     InfoStruct
		wantErr  bool
		absError error
	}{
		{
			name: "case1-file_path_abs_error",
			args: args{
				fd: &descriptor.FileDescriptor{},
			},
			want:     InfoStruct{},
			wantErr:  true,
			absError: fmt.Errorf("error"),
		},
		{
			name: "case2-success",
			args: args{
				fd: &descriptor.FileDescriptor{
					FilePath: "user.proto",
				},
			},
			want: InfoStruct{
				Title:       "user",
				Description: "The api document of user.proto",
				Version:     "2.0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(filepath.Abs,
				func(path string) (string, error) {
					return tt.args.fd.FilePath, tt.absError
				})
			defer p.Reset()

			got, err := NewInfo(tt.args.fd)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
