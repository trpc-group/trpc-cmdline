// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package semver is a library for parsing tool versions.
package semver

import (
	"regexp"
	"strconv"
)

const versionPattern = "(0|(?:[1-9]\\d*))(?:\\.(0|(?:[1-9]\\d*))(?:\\.(0|(?:[1-9]\\d*)))?(?:\\-([\\w][\\w\\.\\-_]*))?)?"

var versionRE = regexp.MustCompile(versionPattern)

// Versions extract the major, minor and revision (patching) version
func Versions(ver string) (major, minor, revision int) {
	var err error
	matches := versionRE.FindStringSubmatch(ver)

	if len(matches) >= 2 {
		major, err = strconv.Atoi(matches[1])
		if err != nil {
			return
		}
	}

	if len(matches) >= 3 {
		minor, err = strconv.Atoi(matches[2])
		if err != nil {
			return
		}
	}

	if len(matches) >= 4 {
		revision, err = strconv.Atoi(matches[3])
		if err != nil {
			return
		}
	}
	return
}

// NewerThan check whether semver `v1` is newer than `v2` or not.
func NewerThan(v1 string, v2 string) bool {
	m1, n1, r1 := Versions(v1)
	m2, n2, r2 := Versions(v2)

	return (m1 > m2) || (m1 == m2 && n1 > n2) || (m1 == m2 && n1 == n2 && r1 > r2)
}
