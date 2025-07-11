// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package completion

import (
	"testing"

	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/cmd/internal"
)

func TestCmd_Completion(t *testing.T) {
	completionCmd := CMD()
	// bash
	output, err := internal.RunAndWatch(completionCmd, nil, []string{"bash"})
	require.Nil(t, err)
	require.NotEmpty(t, output)
	// t.Logf("generated bash completion script: \n%s", output)

	// zsh
	output, err = internal.RunAndWatch(completionCmd, nil, []string{"zsh"})
	require.Nil(t, err)
	require.NotEmpty(t, output)
	// t.Logf("generated zsh completion script: \n%s", output)

	// fish
	output, err = internal.RunAndWatch(completionCmd, nil, []string{"fish"})
	require.Nil(t, err)
	require.NotEmpty(t, output)
	// t.Logf("generated fish completion script: \n%s", output)

	// powershell
	output, err = internal.RunAndWatch(completionCmd, nil, []string{"powershell"})
	require.Nil(t, err)
	require.NotEmpty(t, output)
	// t.Logf("generated powershell completion script: \n%s", output)
}
