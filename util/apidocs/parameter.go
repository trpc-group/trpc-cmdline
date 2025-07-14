// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package apidocs

// ParametersStruct defines the structure of input parameters information for methods in Swagger JSON format.
type ParametersStruct struct {
	Name string `json:"name"` // Name of the parameter.
	// Parameter's passing method: in, query, path, header, form.
	In          string        `json:"in"`
	Required    bool          `json:"required"`              // Whether the parameter is required or not.
	Type        string        `json:"type,omitempty"`        // Type of parameter.
	Schema      *SchemaStruct `json:"schema,omitempty"`      // Parameter reference, optional.
	Format      string        `json:"format,omitempty"`      // Parameter format, optional.
	Description string        `json:"description,omitempty"` // Parameter description, optional.
	Enum        []int32       `json:"enum,omitempty"`        // Possible values for the enum.
	// When type = array, it is necessary to indicate the member type, that is, the description of a single field value.
	Items   *PropertyStruct `json:"items,omitempty"`
	Default interface{}     `json:"default,omitempty"` // Default value of the parameter.
}

// ParameterStructX for v3.
type ParameterStructX struct {
	Name string `json:"name"` // Name of the parameter.
	// Parameter's passing method: in, query, path, header, form.
	In          string       `json:"in"`
	Required    bool         `json:"required"`              // Whether the parameter is required or not.
	Description string       `json:"description,omitempty"` // Parameter description.
	Schema      *ModelStruct `json:"schema,omitempty"`      // Parameter reference.
}

// GetParametersStructX converts the structure.
func (param ParametersStruct) GetParametersStructX() *ParameterStructX {
	ref := ""
	if param.Schema != nil {
		ref = param.Schema.Ref
	}
	return &ParameterStructX{
		Name:        param.Name,
		In:          param.In,
		Required:    param.Required,
		Description: param.Description,
		Schema: &ModelStruct{
			Type:                 param.Type,
			Title:                param.Name,
			Description:          param.Description,
			AdditionalProperties: nil,
			Ref:                  ref,
			Items:                param.Items,
		},
	}
}

// GetRequestBody converts the structure.
func (param ParametersStruct) GetRequestBody() *BodyContentStruct {
	return &BodyContentStruct{
		Content: map[string]MediaStruct{
			"application/json": {
				Schema: *param.Schema,
			},
		},
	}
}

// GetProperty converts the structure.
func (param ParametersStruct) GetProperty() *PropertyStruct {
	ref := ""
	if param.Schema != nil {
		ref = param.Schema.Ref
	}
	return &PropertyStruct{
		Title:       param.Name,
		Type:        param.Type,
		Format:      param.Format,
		Ref:         ref,
		Description: param.Description,
		Enum:        param.Enum,
		Items:       param.Items,
	}
}
