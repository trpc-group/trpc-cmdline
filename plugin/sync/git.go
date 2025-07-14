// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package sync

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// GitManager is the git operations interface.
type GitManager interface {
	PlainInit(path string, isBare bool) (*git.Repository, error)
	PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)

	CreateRemote(r *git.Repository, c *config.RemoteConfig) (*git.Remote, error)
	Remote(r *git.Repository, name string) (*git.Remote, error)
	Worktree(r *git.Repository) (*git.Worktree, error)
	Push(r *git.Repository, o *git.PushOptions) error
	Tags(r *git.Repository) (storer.ReferenceIter, error)
	TagObject(r *git.Repository, h plumbing.Hash) (*object.Tag, error)
	TagObjects(r *git.Repository) (*object.TagIter, error)
	Head(r *git.Repository) (*plumbing.Reference, error)
	CreateTag(r *git.Repository, name string, hash plumbing.Hash,
		opts *git.CreateTagOptions) (*plumbing.Reference, error)

	AddWithOptions(w *git.Worktree, opts *git.AddOptions) error
	Commit(w *git.Worktree, msg string, opts *git.CommitOptions) (plumbing.Hash, error)

	NewPublicKeysFromFile(user, pemFile, password string) (*ssh.PublicKeys, error)
}

type defaultGitManager struct{}

// DefaultGitManager is the constructor of the default Git manager.
var DefaultGitManager = &defaultGitManager{}

// PlainInit create an empty git repository at the given path.
func (d *defaultGitManager) PlainInit(path string, isBare bool) (*git.Repository, error) {
	return git.PlainInit(path, isBare)
}

// PlainClone a repository into the path with the given options.
func (d *defaultGitManager) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}

// CreateRemote creates a new remote
func (d *defaultGitManager) CreateRemote(r *git.Repository, c *config.RemoteConfig) (*git.Remote, error) {
	return r.CreateRemote(c)
}

// Remote return a remote if exists
func (d *defaultGitManager) Remote(r *git.Repository, name string) (*git.Remote, error) {
	return r.Remote(name)
}

// Worktree returns a worktree based on the given fs, if nil the default worktree will be used.
func (d *defaultGitManager) Worktree(r *git.Repository) (*git.Worktree, error) {
	return r.Worktree()
}

// Push performs a push to the remote.
func (d *defaultGitManager) Push(r *git.Repository, o *git.PushOptions) error {
	return r.Push(o)
}

// Tags returns all the References from Tags.
func (d *defaultGitManager) Tags(r *git.Repository) (storer.ReferenceIter, error) {
	return r.Tags()
}

// TagObject returns a Tag with the given hash.
func (d *defaultGitManager) TagObject(r *git.Repository, h plumbing.Hash) (*object.Tag, error) {
	return r.TagObject(h)
}

// TagObjects returns a unsorted TagIter that can step through all of the annotated tags in the repository.
func (d *defaultGitManager) TagObjects(r *git.Repository) (*object.TagIter, error) {
	return r.TagObjects()
}

// Head returns the reference where HEAD is pointing to.
func (d *defaultGitManager) Head(r *git.Repository) (*plumbing.Reference, error) {
	return r.Head()
}

// CreateTag create a tag.
func (d *defaultGitManager) CreateTag(r *git.Repository, name string, hash plumbing.Hash,
	opts *git.CreateTagOptions) (*plumbing.Reference, error) {
	return r.CreateTag(name, hash, opts)
}

// AddWithOptions add with options for worktree.
func (d *defaultGitManager) AddWithOptions(w *git.Worktree, opts *git.AddOptions) error {
	return w.AddWithOptions(opts)
}

// Commit stores the current contents of the index in a new commit along with
// a log message from the user describing the changes.
func (d *defaultGitManager) Commit(w *git.Worktree, msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return w.Commit(msg, opts)
}

// NewPublicKeysFromFile returns a PublicKeys from a file containing a PEM encoded private key.
func (d *defaultGitManager) NewPublicKeysFromFile(user, pemFile, password string) (*ssh.PublicKeys, error) {
	return ssh.NewPublicKeysFromFile(user, pemFile, password)
}
