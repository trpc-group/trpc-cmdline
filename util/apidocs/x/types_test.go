// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package x

import (
	"testing"

	protobuf "google.golang.org/protobuf/types/descriptorpb"
)

func TestGetFormatStr(t *testing.T) {
	tests := []struct {
		name  string
		pType protobuf.FieldDescriptorProto_Type
		want  string
	}{
		{"DOUBLE", protobuf.FieldDescriptorProto_TYPE_DOUBLE, "double"},
		{"FLOAT", protobuf.FieldDescriptorProto_TYPE_FLOAT, "float"},
		{"INT64", protobuf.FieldDescriptorProto_TYPE_INT64, "int64"},
		{"UINT64", protobuf.FieldDescriptorProto_TYPE_UINT64, "uint64"},
		{"INT32", protobuf.FieldDescriptorProto_TYPE_INT32, "int32"},
		{"FIXED64", protobuf.FieldDescriptorProto_TYPE_FIXED64, "fixed64"},
		{"FIXED32", protobuf.FieldDescriptorProto_TYPE_FIXED32, "fixed32"},
		{"BOOL", protobuf.FieldDescriptorProto_TYPE_BOOL, ""},
		{"STRING", protobuf.FieldDescriptorProto_TYPE_STRING, ""},
		{"GROUP", protobuf.FieldDescriptorProto_TYPE_GROUP, "group"},
		{"MESSAGE", protobuf.FieldDescriptorProto_TYPE_MESSAGE, "message"},
		{"BYTES", protobuf.FieldDescriptorProto_TYPE_BYTES, "bytes"},
		{"UINT32", protobuf.FieldDescriptorProto_TYPE_UINT32, "uint32"},
		// Enum type format should be set to int32 (with type as integer)
		{"ENUM", protobuf.FieldDescriptorProto_TYPE_ENUM, "int32"},
		{"SFIXED32", protobuf.FieldDescriptorProto_TYPE_SFIXED32, "sfixed32"},
		{"SFIXED64", protobuf.FieldDescriptorProto_TYPE_SFIXED64, "sfixed64"},
		{"SINT32", protobuf.FieldDescriptorProto_TYPE_SINT32, "sint32"},
		{"SINT64", protobuf.FieldDescriptorProto_TYPE_SINT64, "sint64"},
		{"other set to string", 19, "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFormatStr(tt.pType); got != tt.want {
				t.Errorf("GetFormatStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeStr(t *testing.T) {
	tests := []struct {
		name  string
		pType protobuf.FieldDescriptorProto_Type
		want  string
	}{
		{"bool", protobuf.FieldDescriptorProto_TYPE_BOOL, "boolean"},
		{"double", protobuf.FieldDescriptorProto_TYPE_DOUBLE, "number"},
		{"float", protobuf.FieldDescriptorProto_TYPE_FLOAT, "number"},
		{"int64", protobuf.FieldDescriptorProto_TYPE_INT64, "string"},
		{"int32", protobuf.FieldDescriptorProto_TYPE_INT32, "integer"},
		{"uint64", protobuf.FieldDescriptorProto_TYPE_UINT64, "string"},
		{"uint32", protobuf.FieldDescriptorProto_TYPE_UINT32, "integer"},
		{"fixed64", protobuf.FieldDescriptorProto_TYPE_FIXED64, "string"},
		{"fixed32", protobuf.FieldDescriptorProto_TYPE_FIXED32, "integer"},
		{"sfixed64", protobuf.FieldDescriptorProto_TYPE_SFIXED64, "integer"},
		{"sfixed32", protobuf.FieldDescriptorProto_TYPE_SFIXED32, "integer"},
		{"sint64", protobuf.FieldDescriptorProto_TYPE_SINT64, "integer"},
		{"sint32", protobuf.FieldDescriptorProto_TYPE_SINT32, "integer"},
		// Enum type should be set to integer.
		{"enum", protobuf.FieldDescriptorProto_TYPE_ENUM, "integer"},
		{"string", protobuf.FieldDescriptorProto_TYPE_STRING, "string"},
		{"bytes", protobuf.FieldDescriptorProto_TYPE_BYTES, "string"},
		{"message", protobuf.FieldDescriptorProto_TYPE_MESSAGE, "object"},
		{"other set to string", 19, "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTypeStr(tt.pType); got != tt.want {
				t.Errorf("GetTypeStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
