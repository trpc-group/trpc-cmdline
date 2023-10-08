package sync

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// A mapFileInfo implements fs.FileInfo and fs.DirEntry for a given map file.
type mapFileInfo struct {
	name string
	f    *fstest.MapFile
}

func (i *mapFileInfo) Name() string               { return i.name }
func (i *mapFileInfo) Size() int64                { return int64(len(i.f.Data)) }
func (i *mapFileInfo) Mode() fs.FileMode          { return i.f.Mode }
func (i *mapFileInfo) Type() fs.FileMode          { return i.f.Mode.Type() }
func (i *mapFileInfo) ModTime() time.Time         { return i.f.ModTime }
func (i *mapFileInfo) IsDir() bool                { return i.f.Mode&fs.ModeDir != 0 }
func (i *mapFileInfo) Sys() interface{}           { return i.f.Sys }
func (i *mapFileInfo) Info() (fs.FileInfo, error) { return i, nil }

func newMapFileInfos() map[string]*mapFileInfo {
	res := make(map[string]*mapFileInfo, 3)
	res["/GolandProjects/issue-shoot-owner/test/stub/trpc.group/cpme_contentcreate/test/helloworld"] =
		&mapFileInfo{
			name: "helloworld",
			f: &fstest.MapFile{
				Mode:    fs.ModeDir,
				ModTime: time.Time{},
			},
		}
	res["/GolandProjects/issue-shoot-owner/test/stub/trpc.group/cpme_contentcreate/test/helloworld/hello.pb.go"] =
		&mapFileInfo{
			name: "hello.pb.go",
			f: &fstest.MapFile{
				Mode:    fs.ModePerm,
				ModTime: time.Time{},
			},
		}
	res["/GolandProjects/issue-shoot-owner/test/stub/trpc.group/cpme_contentcreate"+
		"/test/helloworld/hello.trpc.go"] = &mapFileInfo{
		name: "hello.trpc.go",
		f: &fstest.MapFile{
			Mode:    fs.ModePerm,
			ModTime: time.Time{},
		},
	}
	return res
}

func walkDirMockFunc(_ string, fn fs.WalkDirFunc) error {
	for path, de := range newMapFileInfos() {
		if err := fn(path, de, nil); err != nil {
			return err
		}
	}
	return nil
}

func executeGoMock(t *testing.T) (*gomock.Controller, FileManager, GitManager) {
	ctrl := gomock.NewController(t)
	fm := NewMockFileManager(ctrl)
	fm.EXPECT().RemoveAll(gomock.Any()).Return(nil).AnyTimes()
	fm.EXPECT().WalkDir(gomock.Any(), gomock.Any()).DoAndReturn(walkDirMockFunc).AnyTimes()
	fm.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	fm.EXPECT().Open(gomock.Any()).Return(&os.File{}, nil).AnyTimes()
	fm.EXPECT().Create(gomock.Any()).Return(&os.File{}, nil).AnyTimes()
	fm.EXPECT().Close(gomock.Any()).AnyTimes()
	fm.EXPECT().Copy(gomock.Any(), gomock.Any()).Return(int64(1024), nil).AnyTimes()
	fm.EXPECT().UserHomeDir().Return("/Users/veteranchen", nil).AnyTimes()

	gm := NewMockGitManager(ctrl)
	repo, _ := git.Init(memory.NewStorage(), memfs.New())
	remote := git.NewRemote(nil, &config.RemoteConfig{
		URLs: []string{"git@trpc.group:veteranchen/test.git"},
	})
	w, _ := repo.Worktree()
	slice := []*plumbing.Reference{
		plumbing.NewReferenceFromStrings("v1.2.99", "v1.2.99"),
		plumbing.NewReferenceFromStrings("v1.1.2", "v1.1.2"),
		plumbing.NewReferenceFromStrings("v1.1.99", "v1.1.99"),
	}
	iter := storer.NewReferenceSliceIter(slice)
	tag := &object.Tag{
		Name: "v1.2.99",
	}
	tagIter, _ := repo.TagObjects()
	ref := plumbing.NewReferenceFromStrings("12345", "67890")
	gm.EXPECT().PlainInit(gomock.Any(), gomock.Any()).Return(repo, nil).AnyTimes()
	gm.EXPECT().PlainClone(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
			if strings.Contains(path, "initGitRepo") {
				return nil, transport.ErrEmptyRemoteRepository
			}
			if strings.Contains(path, "cloneGitRepoFailure") {
				return nil, transport.ErrInvalidAuthMethod
			}
			return repo, nil
		}).AnyTimes()
	gm.EXPECT().CreateRemote(gomock.Any(), gomock.Any()).Return(remote, nil).AnyTimes()
	gm.EXPECT().Remote(gomock.Any(), gomock.Any()).Return(remote, nil).AnyTimes()
	gm.EXPECT().Worktree(gomock.Any()).Return(w, nil).AnyTimes()
	gm.EXPECT().Push(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	gm.EXPECT().Tags(gomock.Any()).DoAndReturn(func(r *git.Repository) (storer.ReferenceIter, error) {
		return iter, nil
	}).AnyTimes()
	gm.EXPECT().TagObject(gomock.Any(), gomock.Any()).Return(tag, nil).AnyTimes()
	gm.EXPECT().TagObjects(gomock.Any()).Return(tagIter, nil).AnyTimes()
	gm.EXPECT().Head(gomock.Any()).Return(ref, nil).AnyTimes()
	gm.EXPECT().CreateTag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ref, nil).AnyTimes()
	gm.EXPECT().AddWithOptions(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	gm.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).Return(plumbing.NewHash("12345678"), nil).AnyTimes()
	publicKeys, _ := ssh.NewPublicKeys("git", []byte("1234"), "12345")
	gm.EXPECT().NewPublicKeysFromFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(publicKeys, nil).AnyTimes()
	return ctrl, fm, gm
}

func buildGitAndOptions(t *testing.T) (*Git, *params.Option, *descriptor.FileDescriptor, *gomock.Controller) {
	ctrl, fm, gm := executeGoMock(t)
	return NewGit(fm, gm, AuthSupplier), &params.Option{
			Sync:      true,
			Remote:    "",
			NewTag:    true,
			Tag:       "",
			OutputDir: "/GolandProjects/issue-shoot-owner/test/stub/trpc.group/cpme_contentcreate/test/helloworld",
		}, &descriptor.FileDescriptor{
			GoPackage: "trpc.group/veteranchen/test/helloworld",
		}, ctrl
}

func TestSyncGit_Run(t *testing.T) {
	sGit, opts, fd, ctrl := buildGitAndOptions(t)
	defer ctrl.Finish()
	require.Equal(t, "sync_git", sGit.Name())
	require.True(t, sGit.Check(fd, opts))
	err := sGit.Run(fd, opts)
	require.Nil(t, err)

	opts.Tag = "v2.0.99"
	err = sGit.Run(fd, opts)
	require.Nil(t, err)
}

func TestSyncGit_cloneGitDir(t *testing.T) {
	sGit, opts, _, ctrl := buildGitAndOptions(t)
	defer ctrl.Finish()
	dir, r, err := sGit.cloneOrInitGitDir("trpc.group/veteranchen/test/helloworld", opts)
	require.Nil(t, err)
	require.NotNil(t, r)
	require.NotEmpty(t, dir)

	opts.Remote = "https://trpc.group/veteranchen/test.git"
	dir, r, err = sGit.cloneOrInitGitDir("trpc.group/veteranchen/test/helloworld", opts)
	fmt.Println(err)
	require.NotNil(t, err)

	opts.Remote = "git@trpc.group/veteranchen/test.git"
	dir, r, err = sGit.cloneOrInitGitDir("trpc.group/veteranchen/test/helloworld", opts)
	fmt.Println(err)
	require.NotNil(t, err)

	opts.Remote = "git@trpc.group:veteranchen/test"
	dir, r, err = sGit.cloneOrInitGitDir("trpc.group/veteranchen/test/helloworld", opts)
	fmt.Println(err)
	require.NotNil(t, err)
}

func TestSyncGit_Run_initGitRepo(t *testing.T) {
	sGit, opts, fd, ctrl := buildGitAndOptions(t)
	defer ctrl.Finish()
	opts.OutputDir = "/GolandProjects/issue-shoot-owner/test/stub/trpc.group/initGitRepo/test/helloworld"
	err := sGit.Run(fd, opts)
	require.Nil(t, err)
}

func TestSyncGit_Run_cloneGitRepoFailure(t *testing.T) {
	sGit, opts, fd, ctrl := buildGitAndOptions(t)
	defer ctrl.Finish()
	opts.OutputDir = "/GolandProjects/issue-shoot-owner/test/stub/trpc.group/cloneGitRepoFailure/test/helloworld"
	err := sGit.Run(fd, opts)
	require.NotNil(t, err)
}
