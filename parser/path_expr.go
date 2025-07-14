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
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Path expression.
type pathExpression struct {
	LiteralCount int      // number of literal chars (means those not resulting from template variable substitution)
	VarNames     []string // names of parameters (enclosed by {}) in the path
	VarCount     int      // number of named parameters (enclosed by {}) in the path
	Matcher      *regexp.Regexp
	Source       string // Path as defined by the RouteBuilder
	tokens       []string
}

// newPathExpression parses the path and constructs a REST path expression.
func newPathExpression(path string) (*pathExpression, error) {
	expression, literalCount, varNames, varCount, tokens := templateToRE(path)
	compiled, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return &pathExpression{literalCount, varNames, varCount, compiled, expression, tokens}, nil
}

// templateToRE converts a RESTful path to a regular expression.
// See https://github.com/emicklei/go-restful/blob/v3/path_expression_test.go
func templateToRE(template string) (expr string, literalCount int, varNames []string, varCount int, tokens []string) {
	var buffer bytes.Buffer
	buffer.WriteString("^")

	tokens = tokenizePath(template)
	for _, each := range tokens {
		if each == "" {
			continue
		}

		buffer.WriteString("/")
		if !strings.HasPrefix(each, "{") {
			literalCount += len(each)
			encoded := each
			buffer.WriteString(regexp.QuoteMeta(encoded))
			continue
		}

		// check for regular expression in variable
		colon := strings.Index(each, ":")
		var varName string
		if colon != -1 {
			// extract expression
			varName = strings.TrimSpace(each[1:colon])
			paramExpr := strings.TrimSpace(each[colon+1 : len(each)-1])
			if paramExpr == "*" { // special case
				buffer.WriteString("(.*)")
			} else {
				buffer.WriteString(fmt.Sprintf("(%s)", paramExpr)) // between colon and closing moustache
			}
		} else {
			// plain var
			varName = strings.TrimSpace(each[1 : len(each)-1])
			buffer.WriteString("([^/]+?)")
		}
		varNames = append(varNames, varName)
		varCount++
	}
	return strings.TrimRight(buffer.String(), "/") + "(/.*)?$", literalCount, varNames, varCount, tokens
}
