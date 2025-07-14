// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package parser

import (
	"reflect"
	"testing"
)

var tempregexs = []struct {
	template, regex        string
	names                  []string
	literalCount, varCount int
}{
	{"", "^(/.*)?$", nil, 0, 0},
	{"/a/{b}/c/", "^/a/([^/]+?)/c(/.*)?$", []string{"b"}, 2, 1},
	{"/{a}/{b}/{c-d-e}/", "^/([^/]+?)/([^/]+?)/([^/]+?)(/.*)?$", []string{"a", "b", "c-d-e"}, 0, 3},
	{"/{p}/abcde", "^/([^/]+?)/abcde(/.*)?$", []string{"p"}, 5, 1},
	{"/a/{b:*}", "^/a/(.*)(/.*)?$", []string{"b"}, 1, 1},
	{"/a/{b:[a-z]+}", "^/a/([a-z]+)(/.*)?$", []string{"b"}, 1, 1},
}

func TestTemplateToRE(t *testing.T) {
	ok := true
	for i, fixture := range tempregexs {
		actual, lCount, varNames, vCount, _ := templateToRE(fixture.template)
		if actual != fixture.regex {
			t.Logf("regex mismatch, expected:%v , actual:%v, line:%v\n", fixture.regex, actual, i) // 11 = where the data starts
			ok = false
		}
		if lCount != fixture.literalCount {
			t.Logf("literal count mismatch, expected:%v , actual:%v, line:%v\n", fixture.literalCount, lCount, i)
			ok = false
		}
		if vCount != fixture.varCount {
			t.Logf("variable count mismatch, expected:%v , actual:%v, line:%v\n", fixture.varCount, vCount, i)
			ok = false
		}
		if !reflect.DeepEqual(fixture.names, varNames) {
			t.Logf("variable name mismatch, expected:%v , actual:%v, line:%v\n", fixture.names, varNames, i)
			ok = false
		}
	}
	if !ok {
		t.Fatal("one or more expression did not match")
	}
}
