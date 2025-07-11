// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package tpl encapsulates Go's template operations and
// supports generating stub codes and configurations based on template files.
package tpl

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
)

// funcMap is a map of custom template functions used in Go templates.
var funcMap = template.FuncMap{
	// simplify simplifies a Go type based on protobuf type.
	"simplify": lang.PBSimplifyGoType,
	// gopkg returns the Go package name for a protobuf package.
	"gopkg": lang.PBGoPackage,
	// gopkg_simple returns a simplified version of the Go package name.
	"gopkg_simple": lang.PBValidGoPackage,
	// gotype returns the Go type for a protobuf type.
	"gotype": lang.PBGoType,
	// export returns the exported form of an identifier.
	"export": lang.GoExport,
	// gofulltype returns the fully qualified Go type.
	"gofulltype": lang.GoFullyQualifiedType,
	// gofulltypex returns the fully qualified Go type with special characters escaped.
	"gofulltypex": lang.GoFullyQualifiedTypeX,
	// title converts a string to title case.
	"title": lang.Title,
	// untitle converts a title case string to normal case.
	"untitle": lang.UnTitle,
	// trimright trims the right-side characters from a string.
	"trimright": lang.TrimRight,
	// trimleft trims the left-side characters from a string.
	"trimleft": lang.TrimLeft,
	// splitList splits a comma-separated list into an array.
	"splitList": lang.SplitList,
	// Reverse a list.
	"reverse": lang.ReverseList,
	// last returns the last element of an array or slice.
	"last": lang.Last,
	// hasprefix checks if a string has a given prefix.
	"hasprefix": lang.HasPrefix,
	// hassuffix checks if a string has a given suffix.
	"hassuffix": lang.HasSuffix,
	// contains checks if a string contains a given substring.
	"contains": strings.Contains,
	// add adds two integers.
	"add": lang.Add,
	// camelcase converts a string to camel case.
	"camelcase": lang.Camelcase,
	// lowercamelcase converts a string to lower camel case.
	"lowercamelcase": strcase.ToLowerCamel,
	// lower converts a string to lowercase.
	"lower": strings.ToLower,
	// snakecase converts a string to snake case.
	"snakecase": strcase.ToSnake,
	// secvtpl checks if a string is a secure template.
	"secvtpl": lang.CheckSECVTpl,
	// replace replaces all occurrences of a substring with another substring.
	"replace": strings.ReplaceAll,
	// concat concatenates multiple strings.
	"concat": lang.Concat,
	// mergerpc merges RPC paths.
	"mergerpc": lang.MergeRPC,
	// basenamewithoutext returns the base name of a file without the extension.
	"basenamewithoutext": fs.BaseNameWithoutExt,
	"dir":                filepath.Dir,
	"join":               strings.Join,
}
