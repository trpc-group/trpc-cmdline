// Package parser provides the ability of the parser to generate IDC descriptions from specified files.
package parser

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"trpc.group/trpc-go/fbs"

	"github.com/pkg/errors"

	"trpc.group/trpc-go/trpc-cmdline/config"
	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

type idlParser func(protofile string, protodirs []string, opts ...Option) (*descriptor.FileDescriptor, error)

var idlParsers = map[config.IDLType]idlParser{
	config.IDLTypeProtobuf:    ParseProtoFile,
	config.IDLTypeFlatBuffers: ParseFlatbuffers,
}

// Parse parses the given file, which can be either a Protocol Buffer or FlatBuffers file.
func Parse(protofile string, dirs []string, typ config.IDLType, opts ...Option) (*descriptor.FileDescriptor, error) {
	fn, ok := idlParsers[typ]
	if !ok {
		return nil, fmt.Errorf("idltype: %v not supported", typ)
	}

	fd, err := fn(protofile, dirs, opts...)
	if err != nil {
		return nil, fmt.Errorf("parse IDL[%s] %s error: %v", typ, protofile, err)
	}
	return fd, nil
}

// LoadDescriptorSet loads the file descriptor from the given file name.
func LoadDescriptorSet(descriptorSetInFile, protofile string, opts ...Option) (*descriptor.FileDescriptor, error) {
	option := &options{
		aliasOn:  false,
		language: "go",
		rpcOnly:  false,
	}
	for _, o := range opts {
		o(option)
	}
	bytes, err := os.ReadFile(descriptorSetInFile)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile load descriptor_set_in err: %w", err)
	}
	dbpFileDescriptorSet := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(bytes, dbpFileDescriptorSet); err != nil {
		return nil, err
	}
	fileDescriptorMap, err := desc.CreateFileDescriptors(dbpFileDescriptorSet.File)
	if err != nil {
		return nil, err
	}
	d, ok := fileDescriptorMap[protofile]
	if !ok {
		return nil, fmt.Errorf(
			"protofile %s not found in descriptor_set_in file %s", protofile, descriptorSetInFile)
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("os.Getwd err: %w", err)
	}
	return convertFileDescriptor(path.Join(wd, protofile), nil, &descriptor.ProtoFileDescriptor{FD: d}, option)
}

// checkRequirements checks if the requirements are met.
//
// requirements:
// - Must specify fileoption go_package,
//	and the ending part of the go_package must be consistent with the package directive specified package name;
// - packageName must be equal to serviceName,
// 	and if there are multiple service definitions,
// 	only the first service will be processed, and the rest will be ignored;
// - The number of service definitions must not be 0, except for specifying rpconly.

// ParseProtoFile parses a proto file and returns a constructed FileDescriptor object
// that can be used for template filling.
//
// ParseProtoFile is responsible for:
// - parsing the pb file to get the original description information
// - checking project constraints, such as whether custom constraints such as go_option and method option are specified.
func ParseProtoFile(protofile string, protodirs []string, opts ...Option) (*descriptor.FileDescriptor, error) {
	option := &options{
		aliasOn:  false,
		language: "go",
		rpcOnly:  false,
	}
	for _, o := range opts {
		o(option)
	}

	// Parse pb.
	fds, err := parseProtoFile(protofile, protodirs...)
	if err != nil {
		return nil, fmt.Errorf("parseProtoFile err: %+v", err)
	}
	return convertFileDescriptor(protofile, protodirs, &descriptor.ProtoFileDescriptor{FD: fds[0]}, option)
}

func convertFileDescriptor(
	protofile string,
	protodirs []string,
	fd descriptor.Desc,
	option *options,
) (*descriptor.FileDescriptor, error) {
	// Check constraints.
	if err := checkRequirements(fd, option); err != nil {
		return nil, err
	}
	// Construct a FileDescriptor that can be used to guide code generation.
	fileDescriptor := &descriptor.FileDescriptor{FD: fd}
	// Set dependencies (pb files being imported and their output package names)
	mustNilError(fillDependencies(fd, fileDescriptor))
	// Set packageName
	mustNilError(fillPackageName(fd, fileDescriptor))
	// Set fileOptions
	mustNilError(fillFileOptions(fd, fileDescriptor))
	// Set imports
	mustNilError(fillImports(fd, fileDescriptor))
	// Set service
	mustNilError(fillServices(fd, fileDescriptor, option.aliasOn))
	// Set app server
	mustNilError(fillAppServerName(fd, fileDescriptor))
	// SetMessageTypes sets the definitions of the request and response types of the RPC
	mustNilError(fillRPCMessageTypes(fd, fileDescriptor))

	fileDescriptor.RelatvieFilePath = protofile
	if filepath.IsAbs(protofile) {
		fileDescriptor.FilePath = protofile
	} else {
		fp, err := fs.LocateFile(protofile, protodirs)
		if err != nil {
			return nil, fmt.Errorf("fs.LocateFile err: %w", err)
		}
		fileDescriptor.FilePath = fp
	}

	return fileDescriptor, nil
}

// parseProtoFile uses jhump/protoreflect to parse the .proto file and retrieve the file descriptor.
func parseProtoFile(fname string, protodirs ...string) ([]*desc.FileDescriptor, error) {
	parser := protoparse.Parser{
		ImportPaths:           protodirs,
		IncludeSourceCodeInfo: true,
	}
	log.Debug("parseProtoFile: ImportPaths: %+v", protodirs)
	return parser.ParseFiles(fname)
}

func checkRequirements(fd descriptor.Desc, opts *options) error {
	// MUST: service
	if len(fd.GetServices()) == 0 && !opts.rpcOnly {
		return errors.New("service missing")
	}

	if !opts.multiVersion {
		if err := checkMultiVersion(fd); err != nil {
			return err
		}
	}
	return nil
}

var exemptionProtos = []string{
	"trpc.proto",
	"trpc_options.proto",
	"validate.proto",
	"swagger.proto",
	"annotations.proto",
	"http.proto",
}

// checkMultiVersion check if exist multi-version in proto file.
func checkMultiVersion(fd descriptor.Desc) error {
	r, err := regexp.Compile(`^.*/v\d$`)
	if err != nil {
		return err
	}
	for _, el := range loadImports(fd) {
		if matchAny(el.fileName, exemptionProtos) {
			continue
		}
		if ok := r.MatchString(el.importPath); ok {
			return fmt.Errorf(
				"proto: %s, not supported: go_package=\"%s\""+
					"see: trpc --multi-version param",
				el.fileName, el.importPath)
		}
	}

	return nil
}

func matchAny(s string, names []string) bool {
	for i := range names {
		if strings.Contains(s, names[i]) {
			return true
		}
	}
	return false
}

type protodesc struct {
	importPath string
	fileName   string
}

func loadImports(fd descriptor.Desc) []protodesc {
	all := []protodesc{}
	_, importPath := lang.ExplodeImport(fd.GetFileOptions().GetGoPackage())
	all = append(all, protodesc{importPath, fd.GetName()})

	for _, dep := range fd.GetDependencies() {
		all = append(all, loadImports(dep)...)
	}
	return all
}

// ParseFlatbuffers parses a flatbuffers file and returns a FileDescriptor object that can be used for template filling.
func ParseFlatbuffers(protofile string, protodirs []string, opts ...Option) (*descriptor.FileDescriptor, error) {
	option := options{}
	for _, o := range opts {
		o(&option)
	}

	// Parse flatbuffers file.
	parser := fbs.NewParser(protodirs...)
	fds, err := parser.ParseFiles(protofile)
	if err != nil {
		return nil, err
	}
	return convertFileDescriptor(protofile, protodirs, &descriptor.FbsFileDescriptor{FD: fds[0]}, &option)
}

func mustNilError(err error) {
	if err != nil {
		log.Error("error encountered: %v", err)
		os.Exit(1)
	}
}
