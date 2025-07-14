// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package openapi

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs"
)

func TestGenOpenAPI(t *testing.T) {
	type args struct {
		fd     *descriptor.FileDescriptor
		option *params.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		newErr  error
	}{
		{
			name: "case1: new err",
			args: args{
				fd:     &descriptor.FileDescriptor{},
				option: &params.Option{},
			},
			wantErr: true,
			newErr:  fmt.Errorf("err"),
		},
		{
			name: "case1: without err",
			args: args{
				fd:     &descriptor.FileDescriptor{},
				option: &params.Option{},
			},
			wantErr: false,
			newErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := gomonkey.ApplyFunc(
				apidocs.NewOpenAPIJSON,
				func(fd *descriptor.FileDescriptor, option *params.Option) (*apidocs.OpenAPIJSON, error) {
					return &apidocs.OpenAPIJSON{}, tt.newErr
				},
			).ApplyFunc(
				apidocs.WriteJSON,
				func(file string, data interface{}) error {
					return nil
				},
			)

			defer p.Reset()

			if err := GenOpenAPI(tt.args.fd, tt.args.option); (err != nil) != tt.wantErr {
				t.Errorf("GenOpenAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
