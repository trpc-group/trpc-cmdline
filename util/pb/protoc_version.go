// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package pb

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func protocVersion() (version string, err error) {
	// check installed or not
	_, err = exec.LookPath("protoc")
	if err != nil {
		return "", fmt.Errorf("protoc not found, %v", err)
	}

	// print version
	cmd := exec.Command("protoc", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	version = strings.TrimPrefix(strings.TrimSpace(string(output)), "libprotoc ")
	return
}

const (
	requireProtoVersion = "v3.6.0"
)

func isOldProtocVersion() (old bool, err error) {
	version, err := protocVersion()
	if err != nil {
		return
	}
	return oldVersion(version)
}

func oldVersion(version string) (old bool, err error) {
	return !CheckVersionGreaterThanOrEqualTo(version, requireProtoVersion), nil
}

// CheckVersionGreaterThanOrEqualTo check if version meet the requirement
func CheckVersionGreaterThanOrEqualTo(version, required string) bool {
	version = getVersion(version)
	required = getVersion(required)

	m1, n1, r1 := semanticVersion(version)
	m2, n2, r2 := semanticVersion(required)

	if m1 != m2 {
		return m1 > m2
	}
	if n1 != n2 {
		return n1 > n2
	}
	return r1 >= r2
}

func getVersion(version string) string {
	if len(version) != 0 && (version[0] == 'v' || version[0] == 'V') {
		version = version[1:]
	}
	return version
}

// semanticVersion extract the major, minor and revision (patching) version
func semanticVersion(ver string) (major, minor, revision int) {
	vv := strings.Split(ver, ".")

	resultList := make([]int, 3)
	for i := 0; i < len(resultList) && i < len(vv); i++ {
		num, err := strconv.Atoi(vv[i])
		if err != nil {
			break
		}
		resultList[i] = num
	}

	major = resultList[0]
	minor = resultList[1]
	revision = resultList[2]

	return
}
