// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import "sort"

// MethodStruct defines the detailed information of a method in apidocs.
type MethodStruct struct {
	Summary     string                 `json:"summary"`               // Comments of method.
	OperationID string                 `json:"operationId"`           // Name of method.
	Responses   map[string]MediaStruct `json:"responses"`             // Response.
	Parameters  []*ParametersStruct    `json:"parameters"`            // Parameters.
	Tags        []string               `json:"tags"`                  // The service to which the method belongs.
	Description string                 `json:"description,omitempty"` // Description of the method.
}

// Methods is the set of methods.
type Methods struct {
	Elements map[string]*MethodStruct
	Rank     map[string]int
}

// Put inserts an element into the ordered map, and records the element's rank in "Rank".
func (methods *Methods) Put(key string, value *MethodStruct) {
	methods.Elements[key] = value

	if methods.Rank != nil {
		if _, ok := methods.Rank[key]; !ok {
			methods.Rank[key] = len(methods.Elements)
		}
	}
}

// UnmarshalJSON deserializes JSON data.
func (methods *Methods) UnmarshalJSON(b []byte) error {
	return OrderedUnmarshalJSON(b, &methods.Elements, &methods.Rank)
}

// MarshalJSON serializes the method to JSON.
func (methods Methods) MarshalJSON() ([]byte, error) {
	return OrderedMarshalJSON(methods.Elements, methods.Rank)
}

func (methods *Methods) orderedEach(f func(k string, m *MethodStruct)) {
	if methods == nil {
		return
	}

	var keys []string
	for k := range methods.Elements {
		keys = append(keys, k)
	}

	if methods.Rank != nil {
		sort.Slice(keys, func(i, j int) bool {
			return methods.Rank[keys[i]] < methods.Rank[keys[j]]
		})
	} else {
		sort.Strings(keys)
	}

	for _, k := range keys {
		f(k, methods.Elements[k])
	}
}

func (m MethodStruct) refs() []string {
	var refs []string
	for _, responses := range m.Responses {
		if len(responses.Schema.Ref) > 0 {
			refs = append(refs, GetNameByRef(responses.Schema.Ref))
		}
	}
	for _, parameter := range m.Parameters {
		if parameter.Schema != nil && len(parameter.Schema.Ref) > 0 {
			refs = append(refs, GetNameByRef(parameter.Schema.Ref))
		}
		if parameter.Items != nil {
			refs = append(refs, parameter.Items.refs(make(map[string]bool))...)
		}
	}
	return refs
}

// GetMethodX converts MethodStruct to OpenAPI v3 interface.
func (m MethodStruct) GetMethodX() *MethodStructX {
	methodX := &MethodStructX{
		Summary:     m.Summary,
		OperationID: m.OperationID,
		Responses:   make(map[string]BodyContentStruct),
		RequestBody: nil,
		Tags:        m.Tags,
		Description: m.Description,
	}

	// Returned values.
	for status, r := range m.Responses {
		resp := BodyContentStruct{
			Description: r.Description,
			Content: map[string]MediaStruct{
				"application/json": {
					Schema: r.Schema,
				},
			},
		}
		methodX.Responses[status] = resp
	}

	// Parameters.
	var props []*PropertyStruct
	for _, param := range m.Parameters {
		if param.In != "body" {
			methodX.Parameters = append(methodX.Parameters, param.GetParametersStructX())
			continue
		}

		if param.Schema != nil && param.Schema.Ref != "" {
			methodX.RequestBody = param.GetRequestBody()
			continue
		}

		props = append(props, param.GetProperty())
	}

	if len(props) != 0 {
		methodX.RequestBody = &BodyContentStruct{
			Content: map[string]MediaStruct{
				"application/json": {
					Schema: SchemaStruct{
						Properties: props,
					},
				},
			},
		}
	}

	return methodX
}

// GetMethodsX converts MethodStruct to OpenAPI v3 interface.
func (methods Methods) GetMethodsX() MethodsX {
	methodsX := MethodsX{Elements: make(map[string]*MethodStructX)}
	methodsX.Rank = methods.Rank
	methods.orderedEach(func(name string, method *MethodStruct) {
		methodsX.Elements[name] = method.GetMethodX()
	})
	return methodsX
}
