// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package descriptor

import (
	"strings"

	"trpc.group/trpc-go/fbs"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// GoPackage is the key's value of "go package" in file options.
const GoPackage = "go_package"

// FbsFileDescriptor implements the Desc interface and describes all information about a flatbuffers file.
type FbsFileDescriptor struct {
	FD *fbs.SchemaDesc
}

// GetName implements the Desc interface.
func (p *FbsFileDescriptor) GetName() string {
	return p.FD.Name
}

// GetFullyQualifiedName implements the Desc interface.
func (p *FbsFileDescriptor) GetFullyQualifiedName() string {
	return p.FD.Name
}

// GetPackage implements the Desc interface.
func (p *FbsFileDescriptor) GetPackage() string {
	// Flatbuffers support multiple namespaces in a single file
	// Store each namespace in Namespaces according to the order of appearance
	// Here, choose the last appeared namespace as the package name
	return p.FD.Namespaces[len(p.FD.Namespaces)-1]
}

// GetFileOptions implements the Desc interface.
func (p *FbsFileDescriptor) GetFileOptions() FileOpt {
	return NewFbsAttrs(p.FD.Attrs)
}

// GetDependencies implements the Desc interface.
func (p *FbsFileDescriptor) GetDependencies() []Desc {
	var descs []Desc
	for _, dep := range p.FD.Dependencies {
		descs = append(descs, &FbsFileDescriptor{FD: dep})
	}
	return descs
}

// GetServices implements the Desc interface.
func (p *FbsFileDescriptor) GetServices() []ServiceDesc {
	var descs []ServiceDesc
	for _, sd := range p.FD.RPCs {
		descs = append(descs, &FbsServiceDescriptor{SD: sd})
	}
	return descs
}

// GetMessageTypes implements the Desc interface.
func (p *FbsFileDescriptor) GetMessageTypes() []MessageDesc {
	var descs []MessageDesc
	for _, md := range p.FD.Tables {
		descs = append(descs, &FbsMessageDescriptor{MD: md})
	}
	return descs
}

// FbsAttrs implements the FileOpt interface.
// Attrs contains various strings stored in the attribute field of flatbuffers.
// Among them, the following format is customized to provide go package information for flatbuffers files.
//
//	attribute "go_package=trpc.group/trpcprotocol/testapp/testserver"
type FbsAttrs struct {
	Attrs     []string
	KV        map[string]string
	GoPackage *string `protobuf:"bytes,11,opt,name=go_package,json=goPackage" json:"go_package,omitempty"`
}

// NewFbsAttrs creates a new FbsAttrs object.
// During the creation process, it iterates over the attrs string list to generate key-value pairs.
func NewFbsAttrs(attrs []string) *FbsAttrs {
	f := &FbsAttrs{
		Attrs: attrs,
		KV:    make(map[string]string),
	}
	for _, kv := range attrs {
		strs := strings.Split(kv, "=")
		log.Debug("NewFbsAttrs: %v", strs)
		if len(strs) == 2 {
			f.KV[strs[0]] = strs[1]
		}
		if strs[0] == GoPackage {
			f.GoPackage = &strs[1]
		}
		log.Debug("KV: %v", f.KV)
	}
	return f
}

// GetGoPackage implements the FileOpt interface.
func (f *FbsAttrs) GetGoPackage() string {
	if v, ok := f.KV[GoPackage]; ok {
		log.Debug("return: %v", v)
		return v
	}
	log.Debug("return empty string")
	return ""
}

// FbsServiceDescriptor implements the ServiceDesc interface.
// Describes all information of an RPC service.
type FbsServiceDescriptor struct {
	SD *fbs.RPCDesc
}

// GetName implements the ServiceDesc interface.
func (p *FbsServiceDescriptor) GetName() string {
	return p.SD.Name
}

// GetMethods implements the ServiceDesc interface.
func (p *FbsServiceDescriptor) GetMethods() []MethodDesc {
	var descs []MethodDesc
	for _, md := range p.SD.Methods {
		descs = append(descs, &FbsMethodDescriptor{MD: md})
	}
	return descs
}

// FbsMethodDescriptor implements the MethodDesc interface.
type FbsMethodDescriptor struct {
	MD *fbs.MethodDesc
}

// GetName implements the MethodDesc interface.
func (p *FbsMethodDescriptor) GetName() string {
	return p.MD.Name
}

// GetInputType implements the MethodDesc interface.
func (p *FbsMethodDescriptor) GetInputType() MessageDesc {
	return &FbsMessageDescriptor{MD: p.MD.InputTypeDesc}
}

// GetOutputType implements the MethodDesc interface.
func (p *FbsMethodDescriptor) GetOutputType() MessageDesc {
	return &FbsMessageDescriptor{MD: p.MD.OutputTypeDesc}
}

// IsClientStreaming implements the MethodDesc interface.
func (p *FbsMethodDescriptor) IsClientStreaming() bool {
	return p.MD.ClientStreaming
}

// IsServerStreaming implements the MethodDesc interface.
func (p *FbsMethodDescriptor) IsServerStreaming() bool {
	return p.MD.ServerStreaming
}

// GetSourceInfo implements the MethodDesc interface.
func (p *FbsMethodDescriptor) GetSourceInfo() SourceInfo {
	return &FbsSourceInfo{}
}

// FbsMessageDescriptor implements the MessageDesc interface.
type FbsMessageDescriptor struct {
	MD *fbs.TableDesc
}

// GetFile implements the MessageDesc interface.
func (p *FbsMessageDescriptor) GetFile() Desc {
	return &FbsFileDescriptor{FD: p.MD.Schema}
}

// GetFullyQualifiedName implements the MessageDesc interface.
func (p *FbsMessageDescriptor) GetFullyQualifiedName() string {
	if p.MD.Namespace == "" {
		return "." + p.MD.Name
	}
	return "." + p.MD.Namespace + "." + p.MD.Name
}

// FbsSourceInfo implements the SourceInfo interface.
type FbsSourceInfo struct{}

// GetLeadingComments implements the SourceInfo interface.
func (f *FbsSourceInfo) GetLeadingComments() string {
	return ""
}

// GetTrailingComments implements the SourceInfo interface.
func (f *FbsSourceInfo) GetTrailingComments() string {
	return ""
}
