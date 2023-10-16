// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package promptui provides interactive prompts for command-line applications.
package promptui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// WithMinLength limits the minimum input length.
func WithMinLength(minlen int) promptui.ValidateFunc {
	return func(input string) error {
		input = strings.TrimSpace(input)
		if minlen < 0 {
			return nil
		}
		vv := []rune(input)
		if len(vv) < minlen {
			return fmt.Errorf("title must at least %d chars", minlen)
		}
		return nil
	}
}

// Read reads from standard input, specifies the label, the validateFunc, and returns the string.
func Read(label string, validateFunc promptui.ValidateFunc) (string, error) {

	prompt := promptui.Prompt{
		Label:       label,
		Default:     "",
		AllowEdit:   true,
		Validate:    validateFunc,
		Mask:        0,
		HideEntered: false,
		IsVimMode:   false,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}
	return result, nil
}

// ConfirmQuit reads user input to determine whether to quit.
func ConfirmQuit() bool {
	c, err := Read("press q to quit", func(s string) error {
		if len(s) != 0 && len(s) != 1 {
			return errors.New("invalid input")
		}
		return nil
	})
	if err != nil {
		return false
	}
	if c == "q" {
		return true
	}
	return false
}
