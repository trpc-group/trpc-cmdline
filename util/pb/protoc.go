// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

// Package pb encapsulates the protoc execution logic.
package pb

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
	"trpc.group/trpc-go/trpc-cmdline/util/paths"
)

// Constants definition.
const (
	ProtoTRPC              = "trpc/proto/trpc_options.proto"
	ProtoValidate          = "trpc/validate/validate.proto"
	ProtoProtocGenValidate = "validate/validate.proto" // https://github.com/bufbuild/protoc-gen-validate
	ProtoSwagger           = "trpc/swagger/swagger.proto"

	ProtoDir         = "protobuf"
	TrpcImportPrefix = "trpc/"
)

// IsInternalProto tests if `fname` is internal.
func IsInternalProto(fname string) bool {
	if strings.HasPrefix(fname, "google/protobuf") ||
		fname == ProtoTRPC ||
		fname == ProtoValidate ||
		fname == ProtoProtocGenValidate ||
		fname == ProtoSwagger ||
		strings.HasPrefix(fname, TrpcImportPrefix) {
		return true
	}
	return false
}

// Protoc process `protofile` to generate *.pb.go, which is specified by `language`
//
// When using protoc, the following should also be taken into consideration:
// 1. The pb file being processed may import other pb files, e.g., a.proto imports b.proto,
// and the resulting a.pb.go file will call the initialization function in b.pb.go,
// typically named file_${path_to}b_proto_init();
// 2. When executing protoc, the options passed can have an impact on the generated code.
// For instance, protoc -I path-to a.proto and protoc path-to/a.proto will generate different code,
// with the difference being reflected in the function name file${path_to}_b_proto_init().
// The former will generate an empty ${path_to} part, leading to compilation failure.
//
// How to avoid this problem?
//   - The import statements in pb files may contain virtual path information,
//     which needs to be determined based on the search path specified with -I.
//   - When processing pb files with protoc, the pb file names must contain virtual path information.
//     For example, it should be protoc path-to/a.proto instead of protoc -I path-to a.proto.
//   - Additionally, it is essential that there exists a search path in protoc's search path list
//     that is the parent path of the pb file being processed. This is because of how protoc resolves paths.
//
// About the use of optional labels in proto syntax3:
// Compatibility logic needs to be implemented for different versions of protoc:
//   - protoc (~, v3.12.0), pb syntax3 does not support optional labels
//   - protoc [v3.12.0, v3.15.0), pb syntax3 supports optional labels,
//     but the option --experimental_allow_proto3_optional needs to be added.
//   - protoc v3.15.0+, optional labels in pb syntax3 syntax are parsed by default.
//
// ------------------------------------------------------------------------------------------------------------------
//
// Regarding the issue of output paths for pb.go files:
// The paths=source_relative option controls the output filenames, not the import paths.
// The proto compiler associates an import path with each .proto file. When a.proto imports b.proto,
// the import path is used to determine what (if any) import statement to put in a.pb.go. You can set the import paths
// with go_package options in the .proto files, or with --go_opt=M<filename>=<import_path> on the command line.
//
// The proto compiler generates a .pb.go file for each .proto file. There are several ways in which the output
// directory may be determined. For example, if source/a.proto has an import path of example.com/m/foo:
//
// --go_opt=paths=import: Import path; e.g., example.com/m/foo/a.pb.go
// --go_opt=paths=source_relative: Source path; e.g., source/a.pb.go
// --go_opt=module=example.com/m: Path relative to the module flag; e.g., foo/a.pb.go
//
// In the worst case, if none of these suit your needs, you can always generate into a temporary directory and copy the
// file into the desired location. Neither the paths nor module flags have any effect on the contents of the generated
// files.
func Protoc(protodirs []string, protofile, lang, outputdir string, opts ...Option) error {
	options := options{
		pb2ImportPath:  make(map[string]string),
		pkg2ImportPath: make(map[string]string),
	}
	for _, o := range opts {
		o(&options)
	}

	protocArgs, err := genProtocArgs(protodirs, protofile, lang, outputdir, options)
	if err != nil {
		return fmt.Errorf("generate protoc args err: %w", err)
	}

	importPath, ok := options.pb2ImportPath[protofile]
	if ok {
		defer movePbGoFile(protocArgs.argsGoOut, importPath, protocArgs.baseDir, protofile)
	}

	var args []string
	args = append(args, protocArgs.argsProtoPath...)
	args = append(args, protocArgs.argsGoOut)
	args = append(args, protofile)
	if protocArgs.descriptorSetIn != "" {
		args = append(args, protocArgs.descriptorSetIn)
	}

	// pb3 supports "optional" and other labels.
	args, err = makePb3Labels(args)
	if err != nil {
		panic(err)
	}

	return execProtocCommand(args)
}

func genRelPathFromWdWithDirs(protodirs []string, protofile, wd string) (string, error) {
	for _, dir := range protodirs {
		pbPath := filepath.Join(dir, protofile)
		if fin, err := os.Lstat(pbPath); err != nil || fin.IsDir() {
			// If there is an error getting file information or the path is a directory,
			// continue searching for the next one.
			continue
		}

		rel, err := genRelPathFromWd(pbPath, wd)
		if err != nil {
			return "", err
		}

		return rel, nil
	}
	return "", errors.New("no valid relative path found, please check if the file exists")
}

func genRelPathFromWd(protofile, wd string) (string, error) {
	absWd, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("failed to obtain the absolute path for %s: %w", wd, err)
	}
	absPbFile, err := filepath.Abs(protofile)
	if err != nil {
		return "", fmt.Errorf("failed to obtain the absolute path for %s: %w", protofile, err)
	}
	relPath, err := filepath.Rel(absWd, absPbFile)
	if err != nil {
		log.Error("error getting the relative path from %s to %s", absWd, absPbFile)
		return "", fmt.Errorf("Error getting the relative path: %w", err)
	}
	return relPath, nil
}

func execProtocCommand(args []string) error {
	log.Debug("protoc %s", strings.Join(args, " "))

	cmd := exec.Command("protoc", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		msg := `Explicit 'optional' labels are disallowed in the Proto3 syntax`
		str := strings.Join(cmd.Args, " ")
		if strings.Contains(string(output), msg) {
			return fmt.Errorf("run command: `%s`, error: %s...upgrade `protoc` to v3.15.0+", str, string(output))
		}
		return fmt.Errorf("run command: `%s`, error: %s", str, string(output))
	}

	return nil
}

type protocArgs struct {
	baseDir         string
	argsProtoPath   []string
	argsGoOut       string
	descriptorSetIn string
}

func genProtocArgs(protodirs []string, protofile, lang, outputdir string, options options) (*protocArgs, error) {
	baseDir, baseName := filepath.Split(protofile)
	outputdir = strings.TrimSuffix(filepath.Clean(outputdir), "/"+filepath.Clean(baseDir))

	pb2ImportPath := options.pb2ImportPath
	dirs, err := protoSearchDirs(pb2ImportPath)
	if err != nil {
		return nil, fmt.Errorf("proto search dirs err: %w", err)
	}
	protodirs = append(protodirs, dirs...)

	// make --go_out
	argsGoOut := makeProtocOut(pb2ImportPath, lang, outputdir, options)
	args := &protocArgs{baseDir, nil, argsGoOut, ""}
	if options.descriptorSetIn == "" { // --proto_path and --descriptor_set_in cannot coexist.
		p, err := paths.Locate(baseName, protodirs...)
		if err != nil {
			return nil, fmt.Errorf("paths locate err: %w", err)
		}
		p = strings.TrimSuffix(filepath.Clean(p), filepath.Clean(baseDir))
		// make --proto_path
		args.argsProtoPath, err = makeProtoPath(protodirs, p, options)
		if err != nil {
			return nil, fmt.Errorf("make proto path err: %w", err)
		}
		return args, nil
	}
	args.descriptorSetIn = "--descriptor_set_in=" + options.descriptorSetIn
	return args, nil
}

func makePb3Labels(args []string) ([]string, error) {
	// proto3 allow optional
	v, err := protocVersion()
	if err != nil {
		return nil, err
	}
	if CheckVersionGreaterThanOrEqualTo(v, "v3.15.0") {
		// enabled by default, and no longer require --experimental_allow_proto3_optional flag
	} else if CheckVersionGreaterThanOrEqualTo(v, "v3.12.0") {
		// [experimental] adding the "optional" field label, need passing --experimental_allow_proto3_optional flag
		args = append(args, "--experimental_allow_proto3_optional")
	} else if CheckVersionGreaterThanOrEqualTo(v, "v3.6.0") {
		// Not supported, no need for special settings.
	} else {
		// Not supported and version is below the recommended version of trpc.
		log.Info("protoc version too low, please upgrade it")
	}
	return args, nil
}

func movePbGoFile(argsGoOut, pkg, baseDir, protofile string) {
	v := strings.Split(argsGoOut, ":")
	if len(v) != 2 {
		return
	}
	vv := strings.Split(v[1], "stub/")
	if len(vv) != 2 {
		return
	}
	if vv[1] == pkg {
		pdir := filepath.Join(v[1], baseDir)
		target := filepath.Join(pdir, fs.BaseNameWithoutExt(protofile)+".pb.go")
		fs.Move(target, v[1])

		idx := strings.Index(baseDir, "/")
		if idx != -1 {
			path := filepath.Join(v[1], baseDir[0:idx])
			os.RemoveAll(path)
		}
	}
}

func protoSearchDirs(pb2ImportPath map[string]string) ([]string, error) {
	var protodirs []string
	var err error

	// locate trpc.proto
	protodirs, err = trpcProtoSearchDir(pb2ImportPath, protodirs)
	if err != nil {
		return nil, err
	}

	// locate validate.proto
	protodirs, err = validateProtoSearchDir(pb2ImportPath, protodirs)
	if err != nil {
		return nil, err
	}

	// locate protobuf dir
	if isTrpcProtoImported(pb2ImportPath) {
		pbDir, err := paths.Locate(ProtoDir)
		if err == nil {
			// If the trpc system pb directory is found, it is imported, otherwise it is skipped.
			protodirs = append(protodirs, filepath.Join(pbDir, "protobuf"))
		}
	}

	return protodirs, nil
}

func validateProtoSearchDir(pb2ImportPath map[string]string, protodirs []string) ([]string, error) {
	_, secvdep := pb2ImportPath[ProtoValidate]
	if secvdep {
		secvp, err := paths.Locate(ProtoValidate)
		if err != nil {
			return nil, err
		}
		protodirs = append(protodirs, secvp)
	}
	return protodirs, nil
}

func trpcProtoSearchDir(pb2ImportPath map[string]string, protodirs []string) ([]string, error) {
	_, dep := pb2ImportPath[ProtoTRPC]
	if dep {
		p, err := paths.Locate(ProtoTRPC)
		if err != nil {
			return nil, err
		}
		protodirs = append(protodirs, p)
	}
	return protodirs, nil
}

func isTrpcProtoImported(pbpkgMapping map[string]string) bool {
	for k := range pbpkgMapping {
		if strings.HasPrefix(k, TrpcImportPrefix) {
			return true
		}
	}

	return false
}

func makeProtocOut(pb2ImportPath map[string]string, language, outputdir string, options options) string {
	pbpkg := genPbpkg(pb2ImportPath)
	argsGoOut := makeProtocOutByLanguage(language, pbpkg, outputdir)

	if options.validationEnabled {
		_, ok := options.pkg2ImportPath["validate"]
		if ok {
			argsGoOut = fixProtocOut("validate", argsGoOut, language)
		}
	} else if options.secvEnabled {
		_, ok := options.pkg2ImportPath["validate"]
		secvOut := "secv"
		if !ok {
			_, ok = options.pkg2ImportPath["trpc.v2.validate"]
			secvOut = "secv-v2"
		}
		if ok {
			argsGoOut = fixProtocOut(secvOut, argsGoOut, language)
		}
	}
	return argsGoOut
}

func makeProtocOutByLanguage(language string, pbpkg string, outputdir string) string {
	languageToOut := map[string]string{
		"go": fmt.Sprintf("--%s_out=paths=source_relative%s:%s", language, pbpkg, outputdir),
	}

	out := languageToOut[language]
	if len(out) > 0 {
		return out
	}

	// Other unexpected programming languages.
	_ = os.MkdirAll(outputdir, os.ModePerm)
	if len(pbpkg) != 0 {
		pbpkg += ":"
	}
	out = fmt.Sprintf("--%s_out=%s%s", language, pbpkg, outputdir)
	return out
}

func genPbpkg(pb2ImportPath map[string]string) string {
	var pbpkg string

	if len(pb2ImportPath) != 0 {
		for k, v := range pb2ImportPath {

			// 1. The official Google library should be left to protoc and protoc-gen-go to handle.
			// 2. For other imported pb files, if they have the same validGoPkg as the protofile,
			// then the package parsed by protoreflect/jhump for the pb file is empty.
			// To solve the circular dependency problem here!
			if strings.HasPrefix(k, "google/protobuf") || len(v) == 0 {
				continue
			}
			//BUG: protoc-gen-go, https://google.golang.org/protobuf/issues/1151
			//if v == protofileValidGoPkg {
			//	v = "."
			//}
			//if v == protofileValidGoPkg {
			//	continue
			//}
			//pbpkg += ",M" + k + "=" + lang.PBValidGoPackage(v)
			pbpkg += ",M" + k + "=" + v
		}
	}
	return pbpkg
}

func fixProtocOut(secvOut, protocOut, lang string) string {
	new := fmt.Sprintf("--%s_out=lang=%s", secvOut, lang)

	vals := strings.SplitN(protocOut, "=", 2)
	params := vals[1]

	switch lang {
	case "go":
		return new + "," + params
	default:
		return protocOut
	}
}

func makeProtoPath(protodirs []string, must string, options options) ([]string, error) {
	protodirs = append(protodirs, must)
	protodirs = fs.UniqFilePath(protodirs)

	args := []string{}

	// BUG protoc/protoc-gen-go
	// see: https://github.com/golang/protobuf/issues/1252#issuecomment-741626261
	sort.Strings(protodirs)

	wd, _ := os.Getwd()
	for pos, each := range protodirs {
		if wd == each {
			var newProtodirs []string
			newProtodirs = append(newProtodirs, protodirs[0:pos]...)
			newProtodirs = append(newProtodirs, protodirs[pos+1:]...)
			newProtodirs = append(newProtodirs, wd)
			protodirs = newProtodirs
			break
		}
	}

	return genProtoPathArgs(protodirs, args)
}

func genProtoPathArgs(protodirs []string, args []string) ([]string, error) {
	//for _, protodir := range protodirs {
	for i := len(protodirs) - 1; i >= 0; i-- {
		protodir := protodirs[i]
		protodir, err := filepath.Abs(protodir)
		if err != nil {
			continue
		}

		// filter out non-existing directories.
		fin, err := os.Lstat(protodir)
		if err != nil || !fin.IsDir() {
			continue
		}

		args = append(args, fmt.Sprintf("--proto_path=%s", protodir))
	}
	return args, nil
}
