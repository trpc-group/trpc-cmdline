// Package sync uses the post-plugin extension point to push the generated PB stub code to a remote git repository.
package sync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

const (
	sshGitURLPrefix       = "git@"
	sshGitDomainSep       = ":"
	gitURLSuffix          = ".git"
	trpcGeneratedStubName = "stub"
	gitURLPathSep         = "/"
	defaultGitTag         = "v1.1.1"
	gitTagNameSep         = "."
	defaultGitTagLen      = 3
)

// Git provides git synchronization capabilities.
type Git struct {
	auth   transport.AuthMethod
	newErr error

	fileManager FileManager
	gitManager  GitManager
}

// NewGit is a constructor for syncing stubs to a git repository.
// fileManager: File management interface
// gitManager: Git management interface
// supplier: Injection of Git access authorization methods
func NewGit(
	fileManager FileManager,
	gitManager GitManager,
	supplier func(FileManager, GitManager) (transport.AuthMethod, error),
) *Git {
	sg := &Git{
		fileManager: fileManager,
		gitManager:  gitManager,
	}
	sg.auth, sg.newErr = supplier(fileManager, gitManager)
	return sg
}

// AuthSupplier provides access to a git remote repository via SSH by default.
func AuthSupplier(fileManager FileManager, gitManager GitManager) (transport.AuthMethod, error) {
	homePath, err := fileManager.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("os user home dir err: %w", err)
	}
	rsaFilePath := filepath.Join(homePath, ".ssh/id_rsa")
	publicKeys, err := gitManager.NewPublicKeysFromFile("git", rsaFilePath, "")
	if err != nil {
		return nil, fmt.Errorf("git ssh new public keys from file error: %w, file location %s", err, rsaFilePath)
	}
	return publicKeys, nil
}

// Name of the remote git repository sync plugin.
func (s *Git) Name() string {
	return "sync_git"
}

// Check checks whether to perform remote synchronization.
func (s *Git) Check(_ *descriptor.FileDescriptor, opt *params.Option) bool {
	return opt.Sync
}

// Run syncs the remote Git repository.
func (s *Git) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	if s.newErr != nil {
		return s.newErr
	}
	gitDir, r, err := s.cloneOrInitGitDir(fd.GoPackage, opt)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.fileManager.RemoveAll(gitDir); err != nil {
			log.Error("sync git plugin run file manager remove dir:%s error:%v", gitDir, err)
		}
	}()
	outStubDir := filepath.Join(opt.OutputDir, trpcGeneratedStubName)
	if err := s.copyStubToGitDir(outStubDir, gitDir, s.getRemoteGitURLBase(r)); err != nil {
		return err
	}
	return s.commitAndPushGitDir(r, opt)
}

func (s *Git) cloneOrInitGitDir(goPackage string, opt *params.Option) (string, *git.Repository, error) {
	urlPrefix, paths, err := parseGitURLComponent(goPackage, opt)
	if err != nil {
		return "", nil, err
	}
	tempGitDir := filepath.Join(opt.OutputDir, "stub_temp")
	if err := s.fileManager.RemoveAll(tempGitDir); err != nil {
		return "", nil, err
	}
	r, err := s.cloneOrInitGitRepo(urlPrefix+sshGitDomainSep, tempGitDir, paths)
	if err != nil {
		return "", nil, err
	}
	return tempGitDir, r, nil
}

// cloneOrInitGitRepo recursively constructs git url using path.
// For example, the git address of trpc.group/veteranchen/test/helloworld may be:
// trpc.group/veteranchen.git
// trpc.group/veteranchen/test.git
// trpc.group/veteranchen/test/helloworld.git
func (s *Git) cloneOrInitGitRepo(urlPrefix string, tempDir string, paths []string) (*git.Repository, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("not found clone git repository, please create git repository")
	}
	gitURL := urlPrefix + strings.Join(paths, gitURLPathSep) + gitURLSuffix
	r, err := s.gitManager.PlainClone(tempDir, false, &git.CloneOptions{
		Auth:              s.auth,
		URL:               gitURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err == transport.ErrEmptyRemoteRepository {
		return s.initGitRepo(tempDir, gitURL)
	}
	if err == nil && r != nil {
		return r, nil
	}
	return s.cloneOrInitGitRepo(urlPrefix, tempDir, paths[:len(paths)-1])
}

func (s *Git) initGitRepo(tempDir, gitURL string) (*git.Repository, error) {
	r, err := s.gitManager.PlainInit(tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("git init error: %w url: %s", err, tempDir)
	}
	if _, err := s.gitManager.CreateRemote(r, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gitURL},
	}); err != nil {
		return nil, fmt.Errorf("git create remote err: %w gitUrl: %v", err, gitURL)
	}
	return r, nil
}

// parseGitURLComponent parses the given git URL and returns the prefix, as well as the array of path components.
func parseGitURLComponent(goPackage string, opt *params.Option) (string, []string, error) {
	gitURL := opt.Remote
	if gitURL == "" {
		gitURL = sshGitURLPrefix + strings.Replace(goPackage, gitURLPathSep, sshGitDomainSep, 1) + gitURLSuffix
	}
	if !strings.Contains(gitURL, sshGitDomainSep) {
		return "", nil, fmt.Errorf("ssh git url pattern is invalid %s", gitURL)
	}
	if !strings.HasPrefix(gitURL, sshGitURLPrefix) {
		return "", nil, fmt.Errorf("ssh git url pattern is invalid %s", gitURL)
	}
	if !strings.HasSuffix(gitURL, gitURLSuffix) {
		return "", nil, fmt.Errorf("ssh git url pattern is invalid %s", gitURL)
	}
	urlComps := strings.Split(gitURL, sshGitDomainSep)
	return urlComps[0], strings.Split(strings.TrimSuffix(urlComps[1], gitURLSuffix), gitURLPathSep), nil
}

func (s *Git) commitAndPushGitDir(r *git.Repository, opt *params.Option) error {
	w, err := s.gitManager.Worktree(r)
	if err != nil {
		return fmt.Errorf("git repository work tree err: %w", err)
	}
	if err := s.gitManager.AddWithOptions(w, &git.AddOptions{
		All: true,
	}); err != nil {
		return fmt.Errorf("git add err: %w", err)
	}
	if _, err := s.gitManager.Commit(w, "the generated stub pb is pushed to git repository",
		&git.CommitOptions{All: true}); err != nil {
		return fmt.Errorf("git commit err: %w", err)
	}
	// Tag the target repo.
	if opt.NewTag {
		if err = s.setTag(r, opt); err != nil {
			return err
		}
	}
	if err := s.gitManager.Push(r, &git.PushOptions{
		RefSpecs: []config.RefSpec{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"},
		Auth:     s.auth,
	}); err != nil {
		return fmt.Errorf("git push err: %w", err)
	}
	return nil
}

func (s *Git) copyStubToGitDir(localStubDir, localGitDir, gitURLBase string) error {
	pbFiles, err := s.collectPBFileFullPaths(localStubDir)
	if err != nil {
		return err
	}
	pbDir := filepath.Dir(pbFiles[0])
	pbDirBase := filepath.Base(pbDir) // Get the last segment of the directory.
	defaultSuffix := parseDefaultPathSuffix(pbDir, gitURLBase)
	gitDstPath, err := s.extractGitDstDirFullPath(pbDirBase, localGitDir, defaultSuffix)
	if err != nil {
		return err
	}
	for _, f := range pbFiles {
		if err := s.copyPBFileToGitDir(f, gitDstPath); err != nil {
			return fmt.Errorf(
				"copy stub to git local dir error: %w, localStubDir: %s, localGitDir: %s",
				err, localStubDir, localGitDir)
		}
	}
	return nil
}

func parseDefaultPathSuffix(pbDir, gitURLBase string) string {
	dirs := strings.Split(pbDir, string(os.PathSeparator))
	idx := -1
	lastIdx := len(dirs) - 1
	for i := lastIdx; i >= 0; i-- {
		if dirs[i] == gitURLBase {
			idx = i + 1
			break
		}
	}
	if idx == -1 {
		return filepath.Base(pbDir)
	}
	return filepath.Join(dirs[idx:]...)
}

func (s *Git) collectPBFileFullPaths(outDir string) ([]string, error) {
	pbFiles := make([]string, 0, 8)
	if err := s.fileManager.WalkDir(outDir, func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		pbFiles = append(pbFiles, path)
		return nil
	}); err != nil {
		return nil, err
	}
	if len(pbFiles) == 0 {
		return nil, fmt.Errorf("generated stub file is empty")
	}
	if !isSameDir(pbFiles) {
		return nil, fmt.Errorf("generated stub file not is same dir")
	}
	return pbFiles, nil
}

func (s *Git) extractGitDstDirFullPath(pbDirBase, localGitDir, defaultSuffix string) (string, error) {
	dstGitDir := ""
	if err := s.fileManager.WalkDir(localGitDir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && d.IsDir() && d.Name() == pbDirBase {
			dstGitDir = path
		}
		return nil
	}); err != nil {
		return "", err
	}
	if dstGitDir == "" {
		dstGitDir = filepath.Join(localGitDir, defaultSuffix)
		if err := s.fileManager.MkdirAll(dstGitDir, os.ModePerm); err != nil {
			return "", fmt.Errorf("mk git local destination dir %s error: %w", dstGitDir, err)
		}
	}
	return dstGitDir, nil
}

func (s *Git) copyPBFileToGitDir(pbFile string, gitDir string) error {
	dstFile := filepath.Join(gitDir, filepath.Base(pbFile))
	source, err := s.fileManager.Open(pbFile)
	if err != nil {
		return err
	}
	defer s.fileManager.Close(source)
	if err := s.fileManager.RemoveAll(dstFile); err != nil {
		return err
	}
	destination, err := s.fileManager.Create(dstFile)
	if err != nil {
		return err
	}
	defer s.fileManager.Close(destination)
	if _, err := s.fileManager.Copy(destination, source); err != nil {
		return err
	}
	return nil
}

func isSameDir(files []string) bool {
	dir := filepath.Dir(files[0])
	for i := 1; i < len(files); i++ {
		if dir != filepath.Dir(files[i]) {
			return false
		}
	}
	return true
}

func (s *Git) setTag(r *git.Repository, opt *params.Option) error {
	tag := opt.Tag
	if tag == "" {
		var err error
		tag, err = s.evalTagName(r)
		if err != nil {
			return err
		}
	} else {
		if err := s.tagExists(r, tag); err != nil {
			return err
		}
	}
	h, err := s.gitManager.Head(r)
	if err != nil {
		return err
	}
	if _, err := s.gitManager.CreateTag(r, tag, h.Hash(), &git.CreateTagOptions{Message: tag}); err != nil {
		return err
	}
	return nil
}

func (s *Git) tagExists(r *git.Repository, tag string) error {
	tagFoundErr := "tag name is existed"
	tags, err := s.gitManager.TagObjects(r)
	if err != nil {
		return err
	}
	return tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
}

func (s *Git) evalTagName(r *git.Repository) (string, error) {
	tags, err := s.gitManager.Tags(r)
	if err != nil {
		return "", fmt.Errorf("when eval tag name, git get tags err: %w", err)
	}
	lastTag, err := tags.Next()
	for curTag := lastTag; err != io.EOF; curTag, err = tags.Next() {
		lastTag = curTag
	}
	if lastTag == nil {
		return defaultGitTag, nil
	}
	tagObj, err := s.gitManager.TagObject(r, lastTag.Hash())
	if err != nil {
		return "", fmt.Errorf("when eval tag name, git get tag object err: %w", err)
	}
	return genNewTagName(tagObj.Name), nil
}

// genNewTagName generates a new tag name based on the latest tag name.
// If it follows the v1.1.1 format, increment the last version. If it reaches 99, carry over the middle version.
func genNewTagName(lastTagName string) string {
	tags := strings.Split(lastTagName, gitTagNameSep)
	if len(tags) != defaultGitTagLen {
		return defaultGitTag
	}
	v3, err := strconv.Atoi(tags[2])
	if err != nil {
		return defaultGitTag
	}
	v3++
	if v3 < 100 {
		tags[2] = strconv.Itoa(v3)
		return strings.Join(tags, gitTagNameSep)
	}
	v2, err := strconv.Atoi(tags[1])
	if err != nil {
		return defaultGitTag
	}
	v2++
	tags[1], tags[2] = strconv.Itoa(v2), "1"
	return strings.Join(tags, gitTagNameSep)
}

func (s *Git) getRemoteGitURLBase(repo *git.Repository) string {
	r, err := s.gitManager.Remote(repo, "origin")
	if err != nil {
		return ""
	}
	url := r.Config().URLs[0]
	index := strings.LastIndex(url, gitURLPathSep)
	if index == -1 {
		return ""
	}
	return strings.TrimSuffix(url[index+1:], gitURLSuffix)
}
