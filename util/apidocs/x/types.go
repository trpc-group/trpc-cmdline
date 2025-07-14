// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package x provides common utilities for documentation operations.
package x

import protobuf "google.golang.org/protobuf/types/descriptorpb"

// FieldDescriptorProtoTypeFormats is used to convert field type definitions
// from the protobuf library to the openapi data format map.
var FieldDescriptorProtoTypeFormats = map[protobuf.FieldDescriptorProto_Type]string{
	protobuf.FieldDescriptorProto_TYPE_DOUBLE:  "double",
	protobuf.FieldDescriptorProto_TYPE_FLOAT:   "float",
	protobuf.FieldDescriptorProto_TYPE_INT64:   "int64",
	protobuf.FieldDescriptorProto_TYPE_UINT64:  "uint64",
	protobuf.FieldDescriptorProto_TYPE_INT32:   "int32",
	protobuf.FieldDescriptorProto_TYPE_FIXED64: "fixed64",
	protobuf.FieldDescriptorProto_TYPE_FIXED32: "fixed32",
	// Refer to https://swagger.io/specification, boolean type doesn't need to set format.
	protobuf.FieldDescriptorProto_TYPE_BOOL: "",
	// Refer to https://swagger.io/specification, string type doesn't need to set format.
	protobuf.FieldDescriptorProto_TYPE_STRING:   "",
	protobuf.FieldDescriptorProto_TYPE_GROUP:    "group",
	protobuf.FieldDescriptorProto_TYPE_MESSAGE:  "message",
	protobuf.FieldDescriptorProto_TYPE_BYTES:    "bytes",
	protobuf.FieldDescriptorProto_TYPE_UINT32:   "uint32",
	protobuf.FieldDescriptorProto_TYPE_ENUM:     "int32",
	protobuf.FieldDescriptorProto_TYPE_SFIXED32: "sfixed32",
	protobuf.FieldDescriptorProto_TYPE_SFIXED64: "sfixed64",
	protobuf.FieldDescriptorProto_TYPE_SINT32:   "sint32",
	protobuf.FieldDescriptorProto_TYPE_SINT64:   "sint64",
}

// FieldDescriptorProtoTypes is used to convert field types defined in protobuf library to openapi data type map table.
// Since trpc uses jsonpb as the default encoding and decoding method,
// according to the proto specification, int64, uint64, and fixed64 will be serialized as strings.
var FieldDescriptorProtoTypes = map[protobuf.FieldDescriptorProto_Type]string{
	protobuf.FieldDescriptorProto_TYPE_BOOL:     "boolean",
	protobuf.FieldDescriptorProto_TYPE_DOUBLE:   "number",
	protobuf.FieldDescriptorProto_TYPE_FLOAT:    "number",
	protobuf.FieldDescriptorProto_TYPE_INT64:    "string",
	protobuf.FieldDescriptorProto_TYPE_UINT64:   "string",
	protobuf.FieldDescriptorProto_TYPE_INT32:    "integer",
	protobuf.FieldDescriptorProto_TYPE_FIXED64:  "string",
	protobuf.FieldDescriptorProto_TYPE_FIXED32:  "integer",
	protobuf.FieldDescriptorProto_TYPE_UINT32:   "integer",
	protobuf.FieldDescriptorProto_TYPE_SFIXED32: "integer",
	protobuf.FieldDescriptorProto_TYPE_SFIXED64: "integer",
	protobuf.FieldDescriptorProto_TYPE_SINT32:   "integer",
	protobuf.FieldDescriptorProto_TYPE_SINT64:   "integer",
	protobuf.FieldDescriptorProto_TYPE_ENUM:     "integer",
	protobuf.FieldDescriptorProto_TYPE_STRING:   "string",
	protobuf.FieldDescriptorProto_TYPE_BYTES:    "string",
	protobuf.FieldDescriptorProto_TYPE_MESSAGE:  "object",
}

// GetFormatStr returns the specific format of a field based on its protobuf type.
func GetFormatStr(t protobuf.FieldDescriptorProto_Type) string {
	if val, ok := FieldDescriptorProtoTypeFormats[t]; ok {
		return val
	}

	return "string"
}

// GetTypeStr returns the specific field type according to the protobuf type.
func GetTypeStr(t protobuf.FieldDescriptorProto_Type) string {
	if val, ok := FieldDescriptorProtoTypes[t]; ok {
		return val
	}

	return "string"
}
