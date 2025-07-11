// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package fb

// The Option type implements the functional options pattern.
type Option func(*Fbs)

// WithFbsDirs sets the fbsdirs parameter.
func WithFbsDirs(dirs []string) Option {
	return func(o *Fbs) {
		o.args.includePaths = dirs
	}
}

// WithFbsfile sets the fbsfile parameter.
func WithFbsfile(file string) Option {
	return func(o *Fbs) {
		o.args.fbsfile = file
	}
}

// WithLanguage sets the language parameter.
func WithLanguage(language string) Option {
	return func(o *Fbs) {
		o.language = language
	}
}

// WithPackagePath sets the packagePath parameter.
func WithPackagePath(packagePath string) Option {
	return func(o *Fbs) {
		o.packagePath = packagePath
	}
}

// WithOutputdir sets the outputdir parameter.
func WithOutputdir(dir string) Option {
	return func(o *Fbs) {
		o.args.outDir = dir
	}
}

// WithFb2ImportPath sets the mapping relationship between the flatbuffers file names and the import paths.
// For example,
// "./file1.fbs" => "trpc.group/testapp/testserver1"
// "./file2.fbs" => "trpc.group/testapp/testserver2"
func WithFb2ImportPath(m map[string]string) Option {
	return func(o *Fbs) {
		o.fb2ImportPath = m
	}
}

// WithPkg2ImportPath sets the mapping between package name and import path
// For example:
// "trpc.testapp.testserver1" => "trpc.group/testapp/testserver1"
// "trpc.testapp.testserver2" => "trpc.group/testapp/testserver2"
func WithPkg2ImportPath(m map[string]string) Option {
	return func(o *Fbs) {
		o.pkg2ImportPath = m
	}
}
