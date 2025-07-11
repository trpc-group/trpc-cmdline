// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"fmt"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// OpenAPIJSON defines the structure of the JSON file that OpenAPI API documentation needs to load.
type OpenAPIJSON struct {
	OpenAPI string     `json:"openapi"` // Version of OpenAPI.
	Info    InfoStruct `json:"info"`    // Description of the API documentation.

	Paths PathsX `json:"paths"` // Set of specific information for request methods.
	// Definitions of various model data models,
	// including method input and output parameter's structure definitions.
	Components ComponentStruct `json:"components"`
}

// NewOpenAPIJSON returns a new OpenAPIJSON instance.
func NewOpenAPIJSON(fd *descriptor.FileDescriptor, option *params.Option) (*OpenAPIJSON, error) {
	refPrefix = "#/components/schemas/"
	if fd.FD == nil {
		return nil, fmt.Errorf("nil fd")
	}

	defs := NewDefinitions(option, append(fd.FD.GetDependencies(), fd.FD)...)

	// Assemble the information of each method.
	paths, err := NewPaths(fd, option, defs)
	if err != nil {
		return nil, fmt.Errorf("generate openapi error: %w", err)
	}
	pathsX := paths.GetPathsX()

	// Get file's information for assemble header information of Swagger JSON doc.
	info, err := NewInfo(fd)
	if err != nil {
		return nil, err
	}

	openapi := &OpenAPIJSON{
		OpenAPI: "3.0.2",
		Info:    info,
		Paths:   pathsX,
		Components: ComponentStruct{
			Schemas: defs.getUsedModels(paths),
		},
	}

	return openapi, nil
}
