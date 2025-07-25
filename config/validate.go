// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package config

// SupportValidate defines whether a certain language supports validation.
// Validation code will be generated only when:
// 1. The language appears as a key in this map.
// 2. The corresponding value is true.
var SupportValidate = map[string]bool{
	"go": true,
}
