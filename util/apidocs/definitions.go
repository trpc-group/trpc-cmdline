// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package apidocs provides functionality for handling API documentation.
// This package contains structures and methods related to generating API documentation in the
// apidocs JSON format. It includes models, definitions, and various utility functions.
package apidocs

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// ModelStruct defines the structure of the whole data model in the apidocs JSON.
type ModelStruct struct {
	Type                 string          `json:"type"`                           // Data model type
	Properties           *Properties     `json:"properties,omitempty"`           // Data model parameters
	Title                string          `json:"title,omitempty"`                // Data model title
	Description          string          `json:"description,omitempty"`          // Data model description
	AdditionalProperties *PropertyStruct `json:"additionalProperties,omitempty"` // Usage of map value
	Ref                  string          `json:"ref,omitempty"`
	Items                *PropertyStruct `json:"items,omitempty"`
}

// Definitions models
type Definitions struct {
	models map[string]ModelStruct
}

var refPrefix = "#/definitions/"

// GetNameByRef gets the name from a reference.
var GetNameByRef = func(ref string) string {
	return strings.TrimPrefix(ref, refPrefix)
}

// RefName returns a reference with the specified name.
var RefName = func(name string) string {
	return fmt.Sprintf("%s%s", refPrefix, name)
}

// NewDefinitions returns a Definition instance.
func NewDefinitions(option *params.Option, fds ...descriptor.Desc) *Definitions {
	defs := &Definitions{
		models: make(map[string]ModelStruct),
	}

	defs.addModelsByFDs(option, fds...)
	return defs
}

// props gets all properties of the model.
func (m ModelStruct) props() []*PropertyStruct {
	var props []*PropertyStruct

	if m.Properties == nil {
		return props
	}

	m.Properties.orderedEach(func(_ string, v PropertyStruct) {
		prop := v
		props = append(props, &prop)
		if prop.Items != nil {
			props = append(props, prop.Items)
		}
	})

	if m.AdditionalProperties != nil {
		props = append(props, m.AdditionalProperties)
		if m.AdditionalProperties.Items != nil {
			props = append(props, m.AdditionalProperties.Items)
		}
	}
	return props
}

// getUsedModels retrieves used models from paths.
func (defs *Definitions) getUsedModels(paths Paths) map[string]ModelStruct {
	models := make(map[string]ModelStruct)

	var usedRefs []string
	paths.orderedEach(func(_ string, pathsMethod Methods) {
		pathsMethod.orderedEach(func(_ string, method *MethodStruct) {
			usedRefs = append(usedRefs, method.refs()...)
		})
	})
	pos := 0
	searched := make(map[string]bool)
	for {
		if pos >= len(usedRefs) {
			break
		}
		ref := usedRefs[pos]
		pos++

		if !defs.exist(ref) {
			continue
		}
		def := defs.getModel(ref)
		models[ref] = def

		for _, prop := range def.props() {
			usedRefs = append(usedRefs, prop.refs(searched)...)
		}
	}

	return models
}

// getModel retrieves a model by name.
func (defs *Definitions) getModel(name string) ModelStruct {
	return defs.models[name]
}

// addModel adds a model.
func (defs *Definitions) addModel(name string, model ModelStruct) {
	defs.models[name] = model
}

// exist is used to check whether a Model exists.
func (defs *Definitions) exist(name string) bool {
	_, ok := defs.models[name]
	return ok
}

// addAdditionalModel adds an additional model to the Definitions.
// It takes the name and property of the model as input parameters.
func (defs *Definitions) addAdditionalModel(name string, property PropertyStruct) {
	model := ModelStruct{
		Type:                 "object",
		Title:                name,
		AdditionalProperties: &property,
	}

	defs.addModel(name, model)
}

// addModelsByMsg adds models to the Definitions based on the provided message descriptor.
// It takes an option, prefix name, and message descriptor as input parameters.
func (defs *Definitions) addModelsByMsg(option *params.Option, prefixName string, msg *desc.MessageDescriptor) {
	name := prefixName + "." + msg.GetName()
	description := strings.TrimSpace(msg.GetSourceInfo().GetLeadingComments())
	if description == "" {
		description = msg.GetName()
	}
	model := ModelStruct{
		Type:        "object",
		Title:       name,
		Properties:  NewProperties(option, msg, defs),
		Description: description,
	}
	defs.addModel(name, model)

	for _, m := range msg.GetNestedMessageTypes() {
		defs.addModelsByMsg(option, name, m)
	}
}

// addModelsByFDs adds models to the Definitions based on the provided file descriptors.
// It takes an option and one or more file descriptors as input parameters.
func (defs *Definitions) addModelsByFDs(option *params.Option, fds ...descriptor.Desc) {
	for _, fd := range fds {
		for _, msg := range fd.GetMessageTypes() {
			messageDescriptor, ok := msg.(*descriptor.ProtoMessageDescriptor)
			if !ok {
				continue
			}

			defs.addModelsByMsg(option, fd.GetPackage(), messageDescriptor.MD)
		}
	}
}

// getMediaStruct returns a media struct for the given name.
// It takes the name of the struct as an input parameter and returns a map of media struct.
func (defs *Definitions) getMediaStruct(name string) map[string]MediaStruct {
	if !defs.exist(name) {
		return make(map[string]MediaStruct)
	}
	def := defs.getModel(name)
	return map[string]MediaStruct{
		"200": {
			Description: def.Description,
			Schema: SchemaStruct{
				Ref: RefName(name),
			},
		},
	}
}

// filterFields filters the fields of a model based on the provided suffix and fields.
// It takes the name, suffix, and fields as input parameters.
func (defs *Definitions) filterFields(name, suffix string, fields []string) {
	if !defs.exist(name) || len(fields) == 0 {
		return
	}

	def := defs.getModel(name)
	filters := make(map[string][]string)
	for _, f := range fields {
		index := strings.Index(f, ".")
		if index == -1 {
			filters[f] = []string{}
			continue
		}
		filters[f[:index]] = append(filters[f[:index]], f[index+1:])
	}

	newDef := ModelStruct{
		Type:                 def.Type,
		Title:                def.Title + "." + suffix,
		Description:          def.Description,
		AdditionalProperties: def.AdditionalProperties,
	}

	if len(def.Properties.Elements) != 0 {
		newDef.Properties = &Properties{
			Elements: make(map[string]PropertyStruct),
		}
		if def.Properties.Rank != nil {
			newDef.Properties.Rank = make(map[string]int)
		}
	}

	def.Properties.orderedEach(func(name string, p PropertyStruct) {
		fields, ok := filters[name]

		if ok && len(fields) == 0 {
			return
		}

		if len(fields) == 0 {
			newDef.Properties.Put(name, p)
			return
		}

		// There is a next level.
		refName := GetNameByRef(p.Ref)
		defs.filterFields(refName, suffix, fields)
		p.Ref = RefName(refName + "." + suffix)
		if defs.exist(refName + "." + suffix) {
			newDef.Properties.Put(name, p)
		}
	})

	if len(newDef.Properties.Elements) > 0 {
		defs.addModel(newDef.Title, newDef)
	}
}

// getBodyParameters returns the body parameters for the given name.
// It takes the name of the struct as an input parameter and returns an array of body parameters.
func (defs *Definitions) getBodyParameters(name string) []*ParametersStruct {
	if !defs.exist(name) {
		return []*ParametersStruct{}
	}

	return []*ParametersStruct{{
		Name:     "requestBody",
		In:       "body",
		Required: false, // Set as non-required field by default.
		Schema: &SchemaStruct{
			Ref: RefName(name),
		},
	}}
}

// getBodyParameter returns the body parameter for the given name and field.
// It takes the name and field of the struct as input parameters and returns a body parameter.
func (defs *Definitions) getBodyParameter(name, field string) (*ParametersStruct, error) {
	def := defs.getModel(name)
	property, ok := def.Properties.Elements[field]
	if !ok {
		fields := make([]string, 0, len(def.Properties.Elements))
		for k := range def.Properties.Elements {
			fields = append(fields, k)
		}
		return nil, fmt.Errorf(
			"field name %s cannot be found in type %s "+
				"(perhaps what you wrote is type name rather than field name "+
				"or the field has been used by path template), "+
				"the field names available are %+v",
			field, name, fields)
	}
	param := property.GetQueryParameter(field)
	param.In = "body"
	param.Schema = &SchemaStruct{
		Ref:  property.Ref,
		Type: property.Type,
	}
	return param, nil
}

// getQueryParameters returns the query parameters for the given name.
// It takes the name of the struct as an input parameter and returns an array of query parameters.
func (defs *Definitions) getQueryParameters(name string) []*ParametersStruct {
	var params []*ParametersStruct

	if !defs.exist(name) {
		return params
	}

	def := defs.getModel(name)
	def.Properties.orderedEach(func(_ string, prop PropertyStruct) {
		params = append(params, prop.GetQueryParameters(prop.Title, defs, map[string]bool{name: true})...)
	})

	return params
}
