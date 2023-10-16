// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
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

// SwaggerJSON defines the structure of the JSON that contains the swagger API documentation.
type SwaggerJSON struct {
	Swagger string     `json:"swagger"` // Version of Swagger.
	Info    InfoStruct `json:"info"`    // Description information of the API document.

	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`

	// A collection of detailed information for the request method.
	Paths Paths `json:"paths"`
	// Various model definitions, including the structure definitions of method input and output parameters.
	Definitions map[string]ModelStruct `json:"definitions"`
}

// NewSwagger generates swagger documents.
func NewSwagger(fd *descriptor.FileDescriptor, option *params.Option) (*SwaggerJSON, error) {
	refPrefix = "#/definitions/"
	if fd.FD == nil {
		return nil, fmt.Errorf("nil fd")
	}

	// Get the data model obtained from the pb file.
	defs := NewDefinitions(option, append(allDependenciesFds(fd.FD), fd.FD)...)

	// Assemble the information of each method.
	paths, err := NewPaths(fd, option, defs)
	if err != nil {
		return nil, fmt.Errorf("generate swagger error: %w", err)
	}

	// Get file information to assemble Swagger JSON document header information.
	infoMap, err := NewInfo(fd)
	if err != nil {
		return nil, err
	}

	// AssembleSwaggerJSON assembles the complete Swagger JSON document.
	swaggerJSON := &SwaggerJSON{
		Swagger:     "2.0",
		Info:        infoMap,
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       paths,
		Definitions: defs.getUsedModels(paths),
	}
	return swaggerJSON, nil
}

func allDependenciesFds(d descriptor.Desc) []descriptor.Desc {
	deps := d.GetDependencies()
	if len(deps) == 0 {
		return nil
	}

	var allDeps []descriptor.Desc
	allDeps = append(allDeps, deps...)
	for _, dep := range deps {
		allDeps = append(allDeps, allDependenciesFds(dep)...)
	}
	return allDeps
}
