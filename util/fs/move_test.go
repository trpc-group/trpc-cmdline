package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey"
)

func TestMove(t *testing.T) {
	type args struct {
		src string
		dst string
	}

	//backup
	src := filepath.Join(wd, "testcase/move")
	dst := filepath.Join(wd, "testcase/move.bak")
	err := Copy(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(src)
		_ = Rename(dst, src)
	}()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// testcases for move a file
		{"case1.1-src-notexist-dst-notexist", args{"notexist", "move/notexist"}, true},
		{"case1.2-src-notexist-dst-exist", args{"notexist", "move/d"}, true},
		{"case2.1-src-file-dst-notexist-dir(dst)isfolder", args{"move/a", "move/d/a"}, false},
		{"case2.2-src-file-dst-notexist-dir(dst)isnotfolder", args{"move/b", "move/nf/a"}, true},
		{"case2.3-src-file-dst-notexist-dir(dst)notexist", args{"move/b", "move/mf/a"}, true},
		{"case3.1-src-file-dst-folder-dst/basename(src)-exist", args{"move/b", "move/d"}, false},
		{"case3.2-src-file-dst-folder-dst/basename(src)-notexist", args{"move/c", "move/d"}, false},
		{"case4.1-src-file-dst-file", args{"move/fd", "move/fd"}, false},
		// testcases for move a directory
		{"case5.1-src-folder-dst-notexist-dir(dst)folder", args{"move/d", "move/e/d"}, false},
		{"case5.2-src-folder-dst-notexist-dir(dst)file", args{"move/d", "move/nf/d"}, true},
		{"case5.3-src-folder-dst-notexist-dir(dst)notexist", args{"move/e", "move/notexist/e"}, true},
		{"case6.1-src-folder-dst-file", args{"move/e", "move/nf"}, true},
		{"case7.1-src-folder-dst-folder-dst/basename(src)existed+empty", args{"move/z", "move/x/y/"}, false},
		{"case7.2-src-folder-dst-folder-dst/basename(src)existed_notempty", args{"move/z1", "move/x/"}, true},
		{"case7.3-src-folder-dst-folder-dst/basename(src)notexist", args{"move/z1", "move/q"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(wd, "testcase", tt.args.src)
			dst := filepath.Join(wd, "testcase", tt.args.dst)
			if err := Move(src, dst); (err != nil) != tt.wantErr {
				t.Errorf("Move() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isSymLinkToDir(t *testing.T) {
	src := filepath.Join(wd, "testcase/move")
	type args struct {
		symlink string
	}
	tests := []struct {
		name    string
		args    args
		wantYes bool
		wantErr bool
	}{
		{
			name: "case1-exist-file",
			args: args{
				symlink: filepath.Join(src, "a"),
			},
			wantYes: false,
			wantErr: false,
		},
		{
			name: "case2-exist-dir",
			args: args{
				symlink: filepath.Join(src, "d"),
			},
			wantYes: true,
			wantErr: false,
		},
		{
			name: "case3-unknown-path",
			args: args{
				symlink: "unknown_path",
			},
			wantYes: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYes, err := isSymLinkToDir(tt.args.symlink)
			if (err != nil) != tt.wantErr {
				t.Errorf("isSymLinkToDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotYes != tt.wantYes {
				t.Errorf("isSymLinkToDir() gotYes = %v, want %v", gotYes, tt.wantYes)
			}
		})
	}
}

func Test_moveDirectory(t *testing.T) {
	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case1-move_but_dst_not_dir",
			args: args{
				src: filepath.Join(wd, "testcase/move/a"),
				dst: filepath.Join(wd, "testcase/move/b"),
			},
			wantErr: true,
		},
		{
			name: "case2-move_and_dst_not_exist",
			args: args{
				src: filepath.Join(wd, "testcase/move/a"),
				dst: filepath.Join(wd, "testcase/move/d"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(Rename, func(oldpath, newpath string) error {
				return nil
			})
			defer p.Reset()

			if err := moveDirectory(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("moveDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
