// Package params stores the parsed parameter data of the tool.
package params

import (
	"trpc.group/trpc-go/trpc-cmdline/config"
)

// Option command option.
type Option struct {
	// pb/flatbuffers option
	Protodirs    []string // protofile/flatbuffers import path
	Protofile    string   // protofile/flatbuffers file
	ProtofileAbs string   // protofile/flatbuffers absolute path

	UseBaseName bool // Whether to pass protoc/flatc by the basename of "--protofile/--fbs" (default as true)

	// Parses the MethodOption or the "//@alias=" alias in comments to replace the RPC in the .proto file.
	AliasOn bool
	// If enabled, client rpc name in stub will be replaced as alias name, default true.
	AliasAsClientRPCName bool
	PerMethod            bool   // Whether to support splitting files by method.
	OutputDir            string // Project output path.
	Force                bool   // Force write.

	DescriptorSetIn string // Descriptor file specified by "--descriptor_set_in".

	// template option
	Assetdir string         // Service template path.
	Language string         // Development language, such as Go.
	Protocol string         // Protocol type, such as trpc, HTTP, etc.
	IDLType  config.IDLType // IDL file type, such as protobuf, flatbuffers, etc.
	RPCOnly  bool           // Generate only RPC-related code, rather than a complete project.
	// Whether to generate dependent stub code, defaults to false. Only effective when RPCOnly is true.
	DependencyStub bool
	NoGoMod        bool // Do not generate go.mod in the stub code, defaults to false.
	SecvEnabled    bool // SecvEnabled decides whether to enable generation of validation files, default true.

	// gomod option
	GoMod         string // go.mod specified in the current project.
	GoModEx       string // Module extracted from go.mod.
	GoVersion     string // Specify Go version.
	TRPCGoVersion string // Specify trpc-go version.

	// logging option
	Verbose bool // Output verbose log information.

	// Mockgen whether to generate mockgen stub.
	Mockgen bool

	// Gotag custom go tag by protobuf field options.
	Gotag bool

	// KeepOrigRPCName keeps the original RPC name.
	KeepOrigRPCName bool

	// OtherType code generation without IDL files, such as Kafka, HTTP.
	OtherType string

	// Domain name for the generated code address.
	Domain string
	// Group name for the generated code address.
	GroupName string
	// Version suffix for the generated code address (effective for both the main library and dependencies).
	VersionSuffix string

	// Supports importing multiple versions of protocols; for example: xxx/v1/runtime;runtime.
	MultiVersion bool

	// Whether the generated Go stub code's Service Descriptor naming includes the "Service" suffix.
	// Default is to include it.
	NoServiceSuffix bool

	// Parsing the Swagger from MethodOption.
	SwaggerOn bool
	// Use JSON request body instead of query method.
	// Use JSON request body instead of the query way.
	SwaggerOptJSONParam bool
	// Output file name.
	SwaggerOut string

	// Parsing the OpenAPI from MethodOption.
	OpenAPIOn bool
	// Output file name.
	OpenAPIOut string

	// Sort the API documentation according to the order defined in the protobuf.
	OrderByPBName bool

	// Whether to synchronize the Git repository.
	Sync bool
	// If Sync is true, push to the remote Git repository address.
	// The default is the address specified in the "go_package" option of the .proto file.
	Remote string
	// Whether to tag the uploaded repository.
	NewTag bool
	// If Sync is true, set the Git tag. If not specified, the default is in the format "v1.1.1",
	// with the latest tag incremented by 1, and carried over in base 100.
	Tag string
}
