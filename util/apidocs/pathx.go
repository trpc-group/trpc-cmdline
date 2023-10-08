package apidocs

// MethodStructX defines the detailed information of a method in OpenAPI JSON.
type MethodStructX struct {
	Summary     string                       `json:"summary,omitempty"`     // Comment of method.
	OperationID string                       `json:"operationId,omitempty"` // Name of method.
	Responses   map[string]BodyContentStruct `json:"responses,omitempty"`   // Response definition of method.

	// Structure  of method's input parameters except body.
	Parameters  []*ParameterStructX `json:"parameters,omitempty"`
	RequestBody *BodyContentStruct  `json:"requestBody,omitempty"` // Struct definition of method input parameters.

	Tags        []string `json:"tags,omitempty"`        // The service to which the method belongs.
	Description string   `json:"description,omitempty"` // Description of the method.
}

// MethodsX for v3
type MethodsX struct {
	Elements map[string]*MethodStructX
	Rank     map[string]int
}

// UnmarshalJSON deserializes JSON data.
func (method *MethodsX) UnmarshalJSON(b []byte) error {
	return OrderedUnmarshalJSON(b, &method.Elements, &method.Rank)
}

// MarshalJSON serializes to JSON.
func (method MethodsX) MarshalJSON() ([]byte, error) {
	return OrderedMarshalJSON(method.Elements, method.Rank)
}

// PathsX for v3
type PathsX struct {
	Elements map[string]MethodsX
	Rank     map[string]int
}

// UnmarshalJSON deserializes JSON data.
func (paths *PathsX) UnmarshalJSON(b []byte) error {
	return OrderedUnmarshalJSON(b, &paths.Elements, &paths.Rank)
}

// MarshalJSON serializes to JSON.
func (paths PathsX) MarshalJSON() ([]byte, error) {
	return OrderedMarshalJSON(paths.Elements, paths.Rank)
}
