// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

//go:build !windows
// +build !windows

package log

// Color directives.
const (
	ColorReset = "\033[0m"
	ColorGreen = "\033[1;32m"
	ColorRed   = "\033[1;31m"
	ColorPink  = "\033[1;35m"
)
