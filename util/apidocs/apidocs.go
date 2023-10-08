// Package apidocs implements the ability to generate API documentation.
package apidocs

import (
	"encoding/json"
	"os"
)

// ComponentStruct component struct
type ComponentStruct struct {
	Schemas map[string]ModelStruct `json:"schemas"`
}

// BodyContentStruct defines the structure of the response in OpenAPI JSON for a given method.
type BodyContentStruct struct {
	Description string                 `json:"description"` // The description returned by method.
	Content     map[string]MediaStruct `json:"content,omitempty"`
}

// MediaStruct defines the structure of the response in api docs json for a given method.
type MediaStruct struct {
	Description string `json:"description,omitempty"` // The description returned by method.
	// The reference to the data model for the method reference, must have.
	Schema SchemaStruct `json:"schema,omitempty"`
}

// SchemaStruct defines the structure of schema used by data model in api docs json.
type SchemaStruct struct {
	Ref        string            `json:"$ref,omitempty"`
	Properties []*PropertyStruct `json:"properties,omitempty"`
}

// WriteJSON writes JSON.
func WriteJSON(file string, data interface{}) error {
	// Format JSON file, ensure the strings output by json not write in one line.
	jsonByte, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, jsonByte, 0666)
}
