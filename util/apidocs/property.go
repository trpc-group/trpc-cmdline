// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/jhump/protoreflect/desc"
	protobuf "google.golang.org/protobuf/types/descriptorpb"

	"trpc.group/trpc-go/trpc-cmdline/params"

	"trpc.group/trpc-go/trpc-cmdline/util/apidocs/x"
)

// PropertyStruct defines the structure of a single field of a data model in the api docs json.
type PropertyStruct struct {
	Title       string  `json:"title,omitempty"`       // Name of parameter.
	Type        string  `json:"type,omitempty"`        // Type of parameter.
	Format      string  `json:"format,omitempty"`      // Format of parameter.
	Ref         string  `json:"$ref,omitempty"`        // References in parameter.
	Description string  `json:"description,omitempty"` // Description of paramaeter.
	Enum        []int32 `json:"enum,omitempty"`        // Possible values of the enum type.
	// When type is array, specify the member type, i.e., the description of a single field value.
	Items *PropertyStruct `json:"items,omitempty"`
}

// Properties Properties
type Properties struct {
	Elements map[string]PropertyStruct
	Rank     map[string]int
}

// Put puts an element into an ordered map and records the order of the elements in Rank.
func (props *Properties) Put(key string, value PropertyStruct) {
	props.Elements[key] = value

	if props.Rank != nil {
		if _, ok := props.Rank[key]; !ok {
			props.Rank[key] = len(props.Elements)
		}
	}
}

// UnmarshalJSON deserializes JSON data.
func (props *Properties) UnmarshalJSON(b []byte) error {
	return OrderedUnmarshalJSON(b, &props.Elements, &props.Rank)
}

// MarshalJSON serializes to JSON.
func (props Properties) MarshalJSON() ([]byte, error) {
	return OrderedMarshalJSON(props.Elements, props.Rank)
}

// orderedEach traverses in order.
func (props *Properties) orderedEach(f func(k string, prop PropertyStruct)) {
	if props == nil {
		return
	}

	var keys []string
	for k := range props.Elements {
		keys = append(keys, k)
	}

	if props.Rank != nil {
		sort.Slice(keys, func(i, j int) bool {
			return props.Rank[keys[i]] < props.Rank[keys[j]]
		})
	} else {
		sort.Strings(keys)
	}

	for _, k := range keys {
		f(k, props.Elements[k])
	}
}

func (p PropertyStruct) refs(searched map[string]bool) []string {
	if searched[p.Ref] {
		// Circular reference exists.
		return nil
	}

	var refs []string
	if len(p.Ref) > 0 {
		refs = append(refs, GetNameByRef(p.Ref))
		searched[p.Ref] = true
	}

	if p.Items == nil {
		return refs
	}

	return append(refs, p.Items.refs(searched)...)
}

// NewPropertyFunc factory
type NewPropertyFunc func(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct

// NewProperties new
func NewProperties(option *params.Option, msg *desc.MessageDescriptor, defs *Definitions) *Properties {
	if len(msg.GetFields()) == 0 {
		return nil
	}

	// Get message's field information and fill in properties.
	propertiesMap := &Properties{
		Elements: make(map[string]PropertyStruct),
	}

	if option.OrderByPBName {
		propertiesMap.Rank = make(map[string]int)
	}

	for _, field := range msg.GetFields() {
		propertiesMap.Put(field.GetName(), NewProperty(field, defs))
	}

	return propertiesMap
}

// NewProperty new
func NewProperty(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct {
	property := newPropertyFactory(field)(field, defs)

	if property.Ref == "" || len(property.Title) == 0 {
		property.Title = field.GetName()
	}

	if !field.IsRepeated() {
		return property
	}

	p := property
	property.Items = &PropertyStruct{
		Type:   p.Type,
		Format: p.Format,
	}

	if p.Ref != "" {
		property.Items = &PropertyStruct{Ref: p.Ref}
	}
	property.Ref = ""
	property.Type = "array"

	return property
}

func newPropertyFactory(field *desc.FieldDescriptor) NewPropertyFunc {
	isMsg := field.GetType() == protobuf.FieldDescriptorProto_TYPE_MESSAGE

	switch {
	case field.GetType() == protobuf.FieldDescriptorProto_TYPE_ENUM:
		return newEnumProperty
	case field.IsMap():
		return newMapProperty
	case isMsg:
		return newMessageProperty
	default:
		return newBasicProperty
	}
}

func newBasicProperty(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct {
	// Get comment of field.
	var descriptions []string
	field.GetSourceInfo().GetLeadingDetachedComments()
	descriptions = append(descriptions, strings.TrimSpace(field.GetSourceInfo().GetLeadingComments()))
	descriptions = append(descriptions, strings.TrimSpace(field.GetSourceInfo().GetTrailingComments()))

	return PropertyStruct{
		Type:        x.GetTypeStr(field.GetType()),
		Format:      x.GetFormatStr(field.GetType()),
		Description: strings.TrimSpace(strings.Join(descriptions, "\n")),
	}
}

func newEnumProperty(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct {
	property := newBasicProperty(field, defs)
	enums := field.GetEnumType().GetValues()
	if len(enums) == 0 {
		return property
	}

	for _, enum := range enums {
		desc := fmt.Sprintf("%d - %s - %s",
			enum.GetNumber(),
			enum.GetName(),
			strings.TrimSpace(enum.GetSourceInfo().GetTrailingComments()),
		)
		property.Enum = append(property.Enum, enum.GetNumber())
		property.Description += " * " + desc + "\n"
	}

	return property
}

func newMessageProperty(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct {
	return PropertyStruct{
		Ref:         RefName(field.GetMessageType().GetFullyQualifiedName()),
		Description: field.GetSourceInfo().GetLeadingComments() + field.GetSourceInfo().GetTrailingComments(),
	}
}

func newMapProperty(field *desc.FieldDescriptor, defs *Definitions) PropertyStruct {
	name := strings.TrimSuffix(field.GetMessageType().GetFullyQualifiedName(), "entry")

	mapValueField := field.GetMapValueType()
	mapAdditionProperties := PropertyStruct{}
	if mapValueField.GetType() == protobuf.FieldDescriptorProto_TYPE_MESSAGE {
		// For map values that are message types, use reflection to get the actual type
		rMapValue := reflect.ValueOf(mapValueField)
		rProto := reflect.Indirect(rMapValue).FieldByName("proto")
		rTypeName := reflect.Indirect(rProto).FieldByName("TypeName")
		typeName := fmt.Sprint(rTypeName.Elem())[1:]
		mapAdditionProperties.Ref = RefName(typeName)
	} else {
		mapAdditionProperties = newBasicProperty(mapValueField, defs)
	}

	defs.addAdditionalModel(name, mapAdditionProperties)
	return PropertyStruct{
		Type:        x.GetTypeStr(field.GetType()),
		Ref:         RefName(name),
		Description: strings.TrimSpace(field.GetSourceInfo().GetLeadingComments()),
	}
}

// GetQueryParameter converts the given parameters into a query parameter string.
func (p PropertyStruct) GetQueryParameter(name string) *ParametersStruct {
	return &ParametersStruct{
		Required:    false, // Set to non-required field by default.
		Name:        name,
		In:          "query",
		Schema:      nil, // Query parameter should not have schema.
		Type:        p.Type,
		Format:      p.Format,
		Description: p.Description,
		Enum:        p.Enum,
		Items:       p.Items,
	}
}

// GetQueryParameters converts the given parameters into a query parameter string.
func (p PropertyStruct) GetQueryParameters(name string, defs *Definitions,
	searched map[string]bool) []*ParametersStruct {

	var params []*ParametersStruct
	if p.Type != "message" && p.Type != "object" && p.Type != "" {
		return append(params, p.GetQueryParameter(name))
	}
	refName := GetNameByRef(p.Ref)
	def := defs.getModel(refName)
	if searched[refName] {
		return append(params, p.GetQueryParameter(name))
	}
	searched[refName] = true
	def.Properties.orderedEach(func(k string, prop PropertyStruct) {
		params = append(params, prop.GetQueryParameters(name+"."+k, defs, searched)...)
	})
	delete(searched, refName)
	return params
}
