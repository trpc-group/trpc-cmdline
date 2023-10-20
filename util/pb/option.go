// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package pb

type options struct {
	secvEnabled       bool
	validationEnabled bool
	pb2ImportPath     map[string]string
	pkg2ImportPath    map[string]string
	descriptorSetIn   string
}

// Option is used to store the content of the relevant options.
type Option func(*options)

// WithSecvEnabled enables validation and generates stub code using protoc-gen-secv.
// Note: protoc-gen-secv is a modified version of protoc-gen-validate.
// protoc-gen-secv is still not opensourced.
// Please use WithValidationEnabled instead.
func WithSecvEnabled(enabled bool) Option {
	return func(o *options) {
		o.secvEnabled = enabled
	}
}

// WithValidateEnabled enables validation and generates stub code using protoc-gen-validate.
// https://github.com/bufbuild/protoc-gen-validate/tree/v1.0.2
func WithValidateEnabled(enabled bool) Option {
	return func(o *options) {
		o.validationEnabled = enabled
	}
}

// WithPb2ImportPath adds mapping between pb file and import path.
func WithPb2ImportPath(m map[string]string) Option {
	return func(o *options) {
		o.pb2ImportPath = m
	}
}

// WithPkg2ImportPath adds the mapping between package name and import path.
func WithPkg2ImportPath(m map[string]string) Option {
	return func(o *options) {
		o.pkg2ImportPath = m
	}
}

// WithDescriptorSetIn adds the descriptor_set_in option to the command.
func WithDescriptorSetIn(descriptorSetIn string) Option {
	return func(o *options) {
		o.descriptorSetIn = descriptorSetIn
	}
}
