package apidocs

import (
	"fmt"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// OpenAPIJSON defines the structure of the JSON file that OpenAPI API documentation needs to load.
type OpenAPIJSON struct {
	OpenAPI string     `json:"openapi"` // Version of OpenAPI.
	Info    InfoStruct `json:"info"`    // Description of the API documentation.

	Paths PathsX `json:"paths"` // Set of specific information for request methods.
	// Definitions of various model data models,
	// including method input and output parameter's structure definitions.
	Components ComponentStruct `json:"components"`
}

// NewOpenAPIJSON returns a new OpenAPIJSON instance.
func NewOpenAPIJSON(fd *descriptor.FileDescriptor, option *params.Option) (*OpenAPIJSON, error) {
	refPrefix = "#/components/schemas/"
	if fd.FD == nil {
		return nil, fmt.Errorf("nil fd")
	}

	defs := NewDefinitions(option, append(fd.FD.GetDependencies(), fd.FD)...)

	// Assemble the information of each method.
	paths := NewPaths(fd, option, defs)
	pathsX := paths.GetPathsX()

	// Get file's information for assemble header information of Swagger JSON doc.
	info, err := NewInfo(fd)
	if err != nil {
		return nil, err
	}

	openapi := &OpenAPIJSON{
		OpenAPI: "3.0.2",
		Info:    info,
		Paths:   pathsX,
		Components: ComponentStruct{
			Schemas: defs.getUsedModels(paths),
		},
	}

	return openapi, nil
}
