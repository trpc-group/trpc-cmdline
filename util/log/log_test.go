// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	SetVerbose(true)
	require.Equal(t, logVerbose, true)
	Info("log content is: %s", "message")
	Debug("log content is: %s", "message")
	Error("log content is: %s", "message")

	SetVerbose(false)
	require.Equal(t, logVerbose, false)
	Info("log content is: %s", "message")
	Debug("log content is: %s", "message")
	Error("log content is: %s", "message")
}
