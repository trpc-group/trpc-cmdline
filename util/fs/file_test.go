// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package fs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var wd string

func TestMain(m *testing.M) {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = d
	os.Exit(m.Run())
}

func TestBaseFileNameWithoutExt(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1-trpc.proto", args{"trpc.proto"}, "trpc"},
		{"2-hello.world.proto", args{"hello.world.proto"}, "hello.world"},
		{"3-trpc.app.server.go", args{"trpc.app.server.go"}, "trpc.app.server"},
		{"4-trpc.group/group/repo/trpc.app.proto", args{"trpc.group/group/repo/trpc.app.proto"}, "trpc.app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseNameWithoutExt(tt.args.filename); got != tt.want {
				t.Errorf("BaseNameWithoutExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocateProtofileDir(t *testing.T) {

	tests := []struct {
		name     string
		filename string
		search   string
		wantErr  bool
	}{
		{"1-good.dat", "good.dat", filepath.Join(wd, "testcase/a/b/"), false},
		{"2-bad.dat", "bad.dat", filepath.Join(wd, "testcase/a/b/c/"), false},
		{"3-hello.dat", "hello.dat", filepath.Join(wd, "testcase/a/b/c/d/"), false},
		{"4-notexist.dat", "notexit.dat", filepath.Join(wd, "testcase/a/b/c/d/"), true},
		{"5-good.dat", "notexit.dat", filepath.Join(wd, "testcase/"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LocateFile(tt.filename, []string{tt.search})
			if (err != nil) != tt.wantErr {
				t.Errorf("LocateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUniqFilePath(t *testing.T) {
	type args struct {
		dirs []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"testcase-1", args{[]string{"/a", "/b", "/a"}}, []string{"/a", "/b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqFilePath(tt.args.dirs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrepareOutputdir(t *testing.T) {
	type args struct {
		outputdir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"testcase-1", args{filepath.Join(wd, "testcase/a/")}, false},          // target dir already existed, return nil
		{"testcase-2", args{filepath.Join(wd, "testcase/a/b/good.dat")}, true}, // target existed but not dir, return error
		{"testcase-3", args{filepath.Join(wd, "testcase/fff")}, false},         // target dir not existed, create it, return nil
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrepareOutputdir(tt.args.outputdir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareOutputdir() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.name == "testcase-3" {
				os.RemoveAll(tt.args.outputdir)
			}
		})
	}
}
