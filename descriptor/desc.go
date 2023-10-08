// Package descriptor provides the corresponding capabilities for parsing IDL.
package descriptor

import (
	"encoding/json"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

// Desc provides an abstract interface for the descriptors parsed from IDL.
// Currently, two types of IDLs are supported: protobuf and flatbuffers.
type Desc interface {
	// GetName returns the original file name, which may not contain path information or only contain relative paths.
	// For example:
	//   test.pb ./common/test.pb // protobuf
	//   test.fbs ./common/test.fbs // flatbuffers
	GetName() string
	// GetFullyQualifiedName has the same function as GetName.
	GetFullyQualifiedName() string
	// GetPackage returns the defined package name.
	// For protobuf, it is the definition corresponding to the package statement.
	// For flatbuffers, it is the definition corresponding to the namespace statement.
	// For example,
	//   package trpc.testapp.testserver; // protobuf
	//   namespace trpc.testapp.testserver; // flatbuffers
	GetPackage() string
	// GetFileOptions returns the file options defined in the pb file.
	// Usually, it will contain the package name information of the language itself.
	// For example,
	//   // File option related to the go package of protobuf.
	//   option go_package="trpc.group/trpcprotocol/testapp/testserver";
	//   // For flatbuffers, the functions that "file option" supported by extend the "attribute".
	//   attribute "go_package=trpc.group/trpcprotocol/testapp/testserver";
	GetFileOptions() FileOpt
	// GetDependencies returns the descriptor of the files that the current IDL depends on.
	// For protobuf, it is the files imported by the import statement.
	// For flatbuffers, it is the files included by the include statement.
	// For example,
	//   import "common.proto"; // protobuf
	//   include "common.fbs"; // flatbuffers
	GetDependencies() []Desc
	// GetServices returns the descriptor of all the RPC services defined in the file.
	GetServices() []ServiceDesc
	// GetMessageTypes returns the descriptor of all the messages defined in the file.
	GetMessageTypes() []MessageDesc
}

// FileOpt provides an interface for file options.
// "file options" is a term in protobuf.
// This term is kept here for compatibility.
// The corresponding keyword in flatbuffers is "attribute".
// protobuf and flatbuffers have different implementations for this.
type FileOpt interface {
	// GetGoPackage returns the value described by the go_package field.
	// For example,
	//   // File option related to the "go package" of protobuf.
	//   option go_package="trpc.group/trpcprotocol/testapp/testserver";
	//   // For flatbuffers, the functions that "file option" supported by extend the "attribute".
	//   attribute "go_package=trpc.group/trpcprotocol/testapp/testserver";
	GetGoPackage() string
}

// ServiceDesc provides an interface for describing services for different IDLs.
type ServiceDesc interface {
	// GetName returns the name of the RPC service.
	GetName() string
	// GetMethods returns the descriptor of all the methods defined in this RPC service.
	GetMethods() []MethodDesc
}

// MethodDesc provides an interface for describing methods for different IDLs.
type MethodDesc interface {
	// GetName returns the name of the method.
	GetName() string
	// GetInputType returns the descriptor of the "request" type.
	GetInputType() MessageDesc
	// GetOutputType returns the descriptor of the "response" type.
	GetOutputType() MessageDesc
	// IsClientStreaming returns true if it is client streaming.
	IsClientStreaming() bool
	// IsServerStreaming returns true if it is server streaming.
	IsServerStreaming() bool
	// GetSourceInfo returns the comment information of this method in the source file.
	GetSourceInfo() SourceInfo
}

// MessageDesc provides an interface for describing messages in different IDLs.
type MessageDesc interface {
	// GetFile returns the descriptor of the file in which this message is defined.
	GetFile() Desc
	// GetFullyQualifiedName returns the full name of this message.
	// Usually includes package information, such as "trpc.testapp.testserver.TestMessage".
	GetFullyQualifiedName() string
}

// SourceInfo provides an interface for source code comments in different IDLs.
type SourceInfo interface {
	// GetLeadingComments returns the leading comments.
	GetLeadingComments() string
	// GetTrailingComments returns the trailing comments.
	GetTrailingComments() string
}

// ImportDesc provides an interface for parsing information related to "import" part (currently mainly used for Go).
type ImportDesc struct {
	Name string
	Path string
}

// FileDescriptor provides description information for the file scope of an IDL file.
type FileDescriptor struct {
	// FD is an interface for IDL,
	// such as *FileDescriptor parsed from protobuf and *FbsDescriptor parsed from flatbuffers.
	FD Desc

	FilePath         string // Absolute path of the current file
	RelatvieFilePath string // Relatvie path from current executing directory.
	// Extracted from FD, it represents the package name of the IDL (protobuf or flatbuffers), e.g. trpc.$app.$server
	PackageName string
	// GoPackage is extracted from pb fileOption or fbs attribute, such as "trpc.group/.../helloworld"
	GoPackage         string
	BaseGoPackageName string // BaseGoPackageName is usually the last part of GoPackage.
	AppName           string // AppName is extracted from PackageName.
	ServerName        string // ServerName is extracted from PackageName.
	// A pb file may import other pb files.
	// If there are references in the definition of RPC requests and responses,
	// the imported package name corresponding to the type is recorded.
	Imports  []string
	ImportsX []ImportDesc

	FileOptions map[string]interface{} // KV pairs constructed by pb fileOptions or fbs attributes.
	Services    []*ServiceDescriptor   // Supports multiple services, extracted from pb service or fbs rpc_service.
	// The package name that let the key is pb file name and let the value is "protoc".
	Pb2ValidGoPkg map[string]string
	// The importpath in code that let key is pb file name and let value is "go".
	Pb2ImportPath map[string]string
	// pb file <-> its' deps proto files if any
	Pb2DepsPbs map[string][]string

	// The package name that let the key is pb file package directive and let the value is "protoc".
	Pkg2ValidGoPkg map[string]string

	// The importpath in code that let key is pb file package directive and let value is "go".
	Pkg2ImportPath map[string]string

	// RPCMessageType maps message type names to the filename where defined that type.
	RPCMessageType map[string]string // k is pkg.typ defined by pb, v is valid pkg.typ in go.
}

// Dump prints the protobuf file parsing information.
func (fd *FileDescriptor) Dump() {
	log.Debug("************************** FileDescriptor ***********************")
	buf, _ := json.MarshalIndent(fd, "", "  ")
	log.Debug("\n%s", string(buf))
	log.Debug("*****************************************************************")
}

// ValidateEnabled indicates whether the validate check is enabled or not.
func (fd *FileDescriptor) ValidateEnabled() bool {
	for _, k := range fd.Imports {
		if strings.Contains(k, "validate.proto") {
			return true
		}
	}
	return false
}

// ServiceDescriptor provides the description information at the service level.
type ServiceDescriptor struct {
	Name string           // Service name.
	RPC  []*RPCDescriptor // RPC interface definition.
	RPCx []*RPCDescriptor // RPC interface definition, including the RPC alias and original name.
}

// RPCDescriptor provides the description information at the RPC level.
type RPCDescriptor struct {
	Name string // Name of the RPC method.
	Cmd  string // RPC command word.
	// FullyQualifiedCmd is the complete command word used for ServiceDesc and client requests.
	FullyQualifiedCmd string
	// RPC request message type, including package, such as package_a.TypeA
	RequestType string
	// RPC response message type, including package name, such as package_b.TypeB
	ResponseType             string
	LeadingComments          string            // RPC leading comments.
	TrailingComments         string            // RPC trailing comments.
	SwaggerInfo              SwaggerDescriptor // SwaggerDescriptor is used to generate swagger documentation.
	ServerStreaming          bool              // Used to determine if the RPC method is a server-side streaming.
	ClientStreaming          bool              // Used to determine if it's a client-side streaming.
	RequestTypeFileOptions   map[string]interface{}
	ResponseTypeFileOptions  map[string]interface{}
	RequestTypePkgDirective  string
	ResponseTypePkgDirective string
	RESTfulAPIInfo           RESTfulAPIDescriptor // Used for generating stub code related to RESTful APIs.
}

// SwaggerDescriptor is the description information required for generating Swagger API documentation.
type SwaggerDescriptor struct {
	Title       string                             // RPC method name.
	Method      string                             // HTTP's method (if the method supports the HTTP protocol).
	Description string                             // Description of method.
	Params      map[string]*SwaggerParamDescriptor // Description of the RPC parameters
}

// SwaggerParamDescriptor is the description information for Swagger parameters.
type SwaggerParamDescriptor struct {
	Name     string
	Required bool
	Default  string
}

// RESTfulAPIDescriptor is the description information required for generating stub code for RESTful APIs.
type RESTfulAPIDescriptor struct {
	ContentList []*RESTfulAPIContent // RESTful API content.
}

// RESTfulAPIContent is the content of a RESTful API.
type RESTfulAPIContent struct {
	Method       string
	PathTmpl     string
	RequestBody  string
	ResponseBody string
}
