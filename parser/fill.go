package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"sort"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	annotations "trpc.group/trpc/trpc-protocol/pb/go/trpc/api"
	trpc "trpc.group/trpc/trpc-protocol/pb/go/trpc/proto"
	"trpc.group/trpc/trpc-protocol/pb/go/trpc/swagger"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/util/lang"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

const (
	trpcName = "trpc"
)

func fillDependencies(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	// package name, such as "trpc.group/trpcprotocol/testapp/testserver
	pb2ValidGoPkg := make(map[string]string)  // k=pb file name，v=package name processed by protoc
	pkg2ValidGoPkg := make(map[string]string) // k=pb file package directive, v=package name processed by protoc
	pkg2ImportPath := make(map[string]string) // k=pb file package directive, v=importpath in go code
	pb2ImportPath := make(map[string]string)  // k=pb file name，v=importpath in go code
	pb2DepsPbs := make(map[string][]string)
	var err error
	func() {
		// Provide examples for two different cases.
		// 1. No go_package field in file option
		//    validGoPkg = "trpc_testapp_testserver" // Replace '.' with '_'
		//    importPath = "trpc.testapp.testserver" // The value of "package"(protobuf) or "namespace"(flatbuffers).
		// 2. There is a go_package field in file option.
		//    validGoPkg = "testserver" // Get the part after the last forward slash in the "go_package" field.
		//    importPath = "trpc.group/trpcprotocol/testapp/testserver" // Value of the "go_package" field.
		validGoPkg := lang.PBValidGoPackage(fd.GetPackage())
		importPath := fd.GetPackage()
		if opts := fd.GetFileOptions(); opts != nil {
			if gopkgopt := opts.GetGoPackage(); len(gopkgopt) != 0 {
				validGoPkg = lang.PBValidGoPackage(gopkgopt)
				importPath = gopkgopt
				err = multierr.Append(err, checkGoKeyword(fd.GetName(), importPath, validGoPkg))
			}
		}
		pb2ValidGoPkg[fd.GetName()] = validGoPkg
		pb2ImportPath[fd.GetName()] = importPath
	}()
	var f func(descriptor.Desc)
	f = func(fd descriptor.Desc) {
		pbName := fd.GetName()
		pb2DepsPbs[pbName] = []string{}
		for _, dep := range fd.GetDependencies() {
			if len(dep.GetDependencies()) != 0 {
				f(dep)
			} else {
				pb2DepsPbs[dep.GetName()] = []string{}
			}
			fname := dep.GetFullyQualifiedName()
			pkg := dep.GetPackage()
			pkg2ImportPath[pkg] = pkg
			pb2ValidGoPkg[fname] = lang.PBValidGoPackage(pkg)
			var (
				validGoPkg = lang.PBValidGoPackage(pkg)
				importPath = pkg
			)
			if opts := dep.GetFileOptions(); opts != nil {
				if gopkgopt := opts.GetGoPackage(); len(gopkgopt) != 0 {
					validGoPkg = lang.PBValidGoPackage(gopkgopt)
					importPath = gopkgopt
					err = multierr.Append(err, checkGoKeyword(dep.GetName(), importPath, validGoPkg))
				}
			}
			pb2ValidGoPkg[fname] = validGoPkg
			pkg2ImportPath[pkg] = importPath
			pkg2ValidGoPkg[pkg] = validGoPkg
			pb2ImportPath[fname] = importPath
			pb2DepsPbs[pbName] = append(pb2DepsPbs[pbName], fname)
		}
	}
	f(fd)
	nfd.Pb2ValidGoPkg = pb2ValidGoPkg
	nfd.Pkg2ValidGoPkg = pkg2ValidGoPkg
	nfd.Pkg2ImportPath = pkg2ImportPath
	nfd.Pb2ImportPath = pb2ImportPath
	nfd.Pb2DepsPbs = pb2DepsPbs
	return err
}

func checkGoKeyword(file, goPackage, pkgName string) error {
	if token.IsKeyword(pkgName) {
		return fmt.Errorf("please do not use go keyword `%s` as package name in go_package `%s`, file: `%s`",
			pkgName, goPackage, file)
	}
	return nil
}

func fillPackageName(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	nfd.PackageName = fd.GetPackage()
	return nil
}

func fillAppServerName(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	strs := strings.Split(fd.GetPackage(), ".")
	// Needs to meet the package format, i.e. trpc.{appName}.{ServerName}.
	if len(strs) == 3 && strs[0] == trpcName {
		nfd.AppName = strs[1]
		nfd.ServerName = strs[2]
	}
	return nil
}

func fillImports(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	nfd.Imports, nfd.ImportsX = getImports(fd, nfd)
	return nil
}

func fillFileOptions(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	m, err := buildFileOptions(fd.GetFileOptions())
	if err != nil {
		return err
	}
	nfd.FileOptions = m
	if pkg, ok := m["go_package"]; ok {
		var ok bool
		if nfd.GoPackage, ok = pkg.(string); ok {
			importName, _ := lang.ExplodeImport(nfd.GoPackage)
			nfd.BaseGoPackageName = importName
		}
		return nil
	}
	nfd.BaseGoPackageName, _ = lang.ExplodeImport(nfd.PackageName)
	return nil
}

func buildFileOptions(opts descriptor.FileOpt) (map[string]interface{}, error) {
	if opts == nil {
		return nil, nil
	}

	v, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})

	err = json.Unmarshal(v, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func fillServices(fd descriptor.Desc, nfd *descriptor.FileDescriptor, aliasMode bool) error {
	// Traverse all RPC services to fill them into "nfd".
	for _, sd := range fd.GetServices() {
		nsd, err := newServiceDescriptor(fd, sd, aliasMode)
		if err != nil {
			return err
		}
		nfd.Services = append(nfd.Services, nsd)
	}
	return nil
}

func newServiceDescriptor(
	fd descriptor.Desc,
	sd descriptor.ServiceDesc,
	aliasMode bool,
) (*descriptor.ServiceDescriptor, error) {
	nsd := &descriptor.ServiceDescriptor{
		Name: sd.GetName(),
	}
	for _, m := range sd.GetMethods() {
		rpc, rpcxs, err := newRPCDescriptor(fd, sd, m, aliasMode)
		if err != nil {
			return nil, err
		}
		nsd.RPC = append(nsd.RPC, rpc)
		nsd.RPCx = append(nsd.RPCx, rpcxs...)
	}
	if err := checkRESTfulAPIInfo(nsd); err != nil {
		return nil, err
	}
	return nsd, nil
}

func newRPCDescriptor(
	fd descriptor.Desc,
	sd descriptor.ServiceDesc,
	m descriptor.MethodDesc,
	aliasMode bool,
) (rpc *descriptor.RPCDescriptor, rpcxs []*descriptor.RPCDescriptor, err error) {
	reqOpts, err := buildFileOptions(m.GetInputType().GetFile().GetFileOptions())
	if err != nil {
		return nil, nil, err
	}
	rspOpts, err := buildFileOptions(m.GetOutputType().GetFile().GetFileOptions())
	if err != nil {
		return nil, nil, err
	}
	rpc = &descriptor.RPCDescriptor{
		Name:              m.GetName(),
		Cmd:               m.GetName(),
		FullyQualifiedCmd: fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), sd.GetName(), m.GetName()),
		RequestType:       m.GetInputType().GetFullyQualifiedName(),
		ResponseType:      m.GetOutputType().GetFullyQualifiedName(),
		LeadingComments: strings.Replace( //compatible with go1.11
			strings.TrimSpace(m.GetSourceInfo().GetLeadingComments()), "\n", "\n// ", -1),
		TrailingComments: strings.Replace(
			strings.TrimSpace(m.GetSourceInfo().GetTrailingComments()), "\n", "\n// ", -1),
		ClientStreaming:          m.IsClientStreaming(),
		ServerStreaming:          m.IsServerStreaming(),
		RequestTypePkgDirective:  m.GetInputType().GetFile().GetPackage(),
		ResponseTypePkgDirective: m.GetOutputType().GetFile().GetPackage(),
		RequestTypeFileOptions:   reqOpts,
		ResponseTypeFileOptions:  rspOpts,
	}

	pmd, ok := m.(*descriptor.ProtoMethodDescriptor)
	if !ok {
		return rpc, rpcxs, nil
	}

	if alias, ok := parseAliasExtension(pmd.MD.GetMethodOptions()); ok {
		rpc := *rpc
		rpc.FullyQualifiedCmd = alias
		rpcxs = append(rpcxs, &rpc)
	}

	if aliasMode {
		alias, ok, err := parseAliasComment(rpc.LeadingComments, rpc.TrailingComments)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			rpc := *rpc
			rpc.FullyQualifiedCmd = alias
			rpcxs = append(rpcxs, &rpc)
		}
	}

	rpc.SwaggerInfo = *parseSwaggerInfo(pmd.MD, rpc.LeadingComments)

	contents, err := parseRestAPIContents(pmd.MD)
	if err != nil {
		return nil, nil, err
	}
	rpc.RESTfulAPIInfo.ContentList = append(rpc.RESTfulAPIInfo.ContentList, contents...)

	return rpc, rpcxs, nil
}

func parseSwaggerInfo(md *desc.MethodDescriptor, leadingComments string) *descriptor.SwaggerDescriptor {
	// If the title is empty, the leading comment defined in RPC will be taken as the title of the method.
	altTile := strings.Replace(leadingComments, "\n", "\n// ", -1)
	if si, ok := parseSwaggerV1(md.GetMethodOptions(), altTile); ok {
		return si
	}
	return parseSwaggerDefault(altTile)
}

func parseSwaggerDefault(altTitle string) *descriptor.SwaggerDescriptor {
	return &descriptor.SwaggerDescriptor{
		Title:       altTitle,
		Method:      "post",
		Description: "",
		Params:      make(map[string]*descriptor.SwaggerParamDescriptor),
	}
}

func parseSwaggerV1(m protoreflect.ProtoMessage, altTitle string) (*descriptor.SwaggerDescriptor, bool) {
	if ok := proto.HasExtension(m, swagger.E_Swagger); !ok {
		return nil, false
	}
	v := proto.GetExtension(m, swagger.E_Swagger)
	sr, ok := v.(*swagger.SwaggerRule)
	if !ok {
		return nil, false
	}
	return getSwaggerInfo[*swagger.SwaggerParam](sr, altTitle), true
}

func getSwaggerInfo[
	SP SwaggerParam, SR SwaggerRule[SP],
](swagger SR, altTitle string) *descriptor.SwaggerDescriptor {
	var si descriptor.SwaggerDescriptor
	if title := strings.TrimSpace(swagger.GetTitle()); len(title) == 0 {
		si.Title = altTitle
	} else {
		si.Title = title
	}
	si.Description = strings.TrimSpace(swagger.GetDescription())
	if method := strings.TrimSpace(swagger.GetMethod()); len(method) == 0 {
		si.Method = "post" // default: POST
	} else {
		si.Method = method
	}

	si.Params = make(map[string]*descriptor.SwaggerParamDescriptor)
	for _, param := range swagger.GetParams() {
		si.Params[param.GetName()] = &descriptor.SwaggerParamDescriptor{
			Name:     param.GetName(),
			Required: param.GetRequired(),
			Default:  param.GetDefault(),
		}
	}
	return &si
}

func parseRestAPIContents(m *desc.MethodDescriptor) ([]*descriptor.RESTfulAPIContent, error) {
	mo := m.GetMethodOptions()
	if proto.HasExtension(mo, annotations.E_Http) {
		if httpRule, ok := proto.GetExtension(mo, annotations.E_Http).(*annotations.HttpRule); ok {
			return parseRestContents(httpRule)
		}
	}
	return nil, nil
}

func parseRestContents[HR HttpRule[HR]](httpRule HR) ([]*descriptor.RESTfulAPIContent, error) {
	var contents []*descriptor.RESTfulAPIContent
	for _, hr := range append([]HR{httpRule}, expandAdditionalBindings(httpRule)...) {
		content := getRESTfulAPIContent(hr)
		if content == nil {
			return nil, fmt.Errorf("get restful api content error")
		}
		contents = append(contents, content)
	}
	return contents, nil
}

func expandAdditionalBindings[HR HttpRule[HR]](rule HR) []HR {
	var rs []HR
	for _, r := range rule.GetAdditionalBindings() {
		rs = append(rs, r)
		rs = append(rs, expandAdditionalBindings(r)...)
	}
	return rs
}

func getRESTfulAPIContent[HR HttpRule[HR]](httpRule HR) *descriptor.RESTfulAPIContent {
	method, pathTmpl, err := parseRestMethodPathTmpl(httpRule)
	if err != nil {
		return nil
	}
	return &descriptor.RESTfulAPIContent{
		Method:       method,
		PathTmpl:     pathTmpl,
		RequestBody:  httpRule.GetBody(),
		ResponseBody: httpRule.GetResponseBody(),
	}
}

func parseRestMethodPathTmpl[HR HttpRule[HR]](hr HR) (string, string, error) {
	switch hr := interface{}(hr).(type) {
	case *annotations.HttpRule:
		return parseRestMethodPathTmplV1(hr)
	default:
		panic("never happen")
	}
}

func parseRestMethodPathTmplV1(hr *annotations.HttpRule) (string, string, error) {
	switch p := hr.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "GET", p.Get, nil
	case *annotations.HttpRule_Put:
		return "PUT", p.Put, nil
	case *annotations.HttpRule_Post:
		return "POST", p.Post, nil
	case *annotations.HttpRule_Delete:
		return "DELETE", p.Delete, nil
	case *annotations.HttpRule_Patch:
		return "PATCH", p.Patch, nil
	case *annotations.HttpRule_Custom:
		return p.Custom.Kind, p.Custom.Path, nil
	default:
		log.Error("unknown RESTful httpRule: %T", hr.Pattern)
		return "", "", fmt.Errorf("unknown RESTful httpRule: %T", hr.Pattern)
	}
}

func checkRESTfulAPIInfo(nsd *descriptor.ServiceDescriptor) error {
	var rpcList []*descriptor.RPCDescriptor
	rpcList = append(rpcList, nsd.RPC...)

	// Get the RESTful Content list.
	restfulContentList := getRESTfulContentList(rpcList)

	pathSet := make(map[string]bool)
	for _, each := range restfulContentList {

		// Check if the resource is correct.
		if _, err := newPathExpression(each.PathTmpl); err != nil {
			return fmt.Errorf("invalid RESTful http path: %s, parse error: %v", each.PathTmpl, err)
		}

		key := fmt.Sprintf("%s:%s", each.Method, each.PathTmpl)
		if pathSet[key] {
			// Duplicate RESTful resources
			return fmt.Errorf("exist repeated RESTful http path:%s", each)
		}
		pathSet[key] = true
	}

	return nil
}

// Tokenize a URL path using the slash separator ; the result does not have empty tokens
func tokenizePath(path string) []string {
	if "/" == path {
		return nil
	}
	return strings.Split(strings.Trim(path, "/"), "/")
}

func getRESTfulContentList(rpcList []*descriptor.RPCDescriptor) []*descriptor.RESTfulAPIContent {
	var restfulContentList []*descriptor.RESTfulAPIContent
	for _, rpc := range rpcList {
		if rpc == nil {
			continue
		}

		restfulContentList = append(restfulContentList, rpc.RESTfulAPIInfo.ContentList...)
	}
	return restfulContentList
}

func parseAliasExtension(m protoreflect.ProtoMessage) (string, bool) {
	aliasName, ok := gatherAliasExtension(m, trpc.E_Alias)
	if ok {
		return aliasName, true
	}
	return "", false
}

func gatherAliasExtension(
	m protoreflect.ProtoMessage,
	et protoreflect.ExtensionType,
) (string, bool) {
	ok := proto.HasExtension(m, et)
	if !ok {
		return "", false
	}
	v := proto.GetExtension(m, et)
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	if aliasName := strings.TrimSpace(s); len(aliasName) != 0 {
		return aliasName, true
	}
	return "", false
}

func parseAliasComment(leading, trailing string) (string, bool, error) {
	alias, err := parseComment(leading, trailing)
	if err == errAnnotationNotFound {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return alias, alias != "", nil
}

// fillRPCMessageTypes: In the stub code,
// mapping relationships between RPC request/response type names and their corresponding Protobuf definitions
// need to be established.
func fillRPCMessageTypes(fd descriptor.Desc, nfd *descriptor.FileDescriptor) error {
	def := make(map[string]string)

	for _, sd := range fd.GetServices() {
		for _, m := range sd.GetMethods() {
			if err := fillRPCMessageTypesByMethod(fd, m, def); err != nil {
				return err
			}
		}
	}

	if len(def) != 0 {
		nfd.RPCMessageType = def
	}
	return nil
}

func fillRPCMessageTypesByMethod(fd descriptor.Desc, m descriptor.MethodDesc, def map[string]string) error {
	in := m.GetInputType().GetFullyQualifiedName()
	out := m.GetOutputType().GetFullyQualifiedName()

	inDefLoc, err := findMessageDefLocation(in, fd)
	if err != nil {
		return err
	}
	def[in] = inDefLoc

	outDefLoc, err := findMessageDefLocation(out, fd)
	if err != nil {
		return err
	}
	def[out] = outDefLoc
	return nil
}

func findMessageDefLocation(typ string, fd descriptor.Desc) (string, error) {
	if s, done := findMessageDefLocationFromMessageType(typ, fd); done {
		return s, nil
	}

	if s, done := findMessageDefLocationFromDependencies(typ, fd); done {
		return s, nil
	}

	return "", errors.New("not found")
}

func findMessageDefLocationFromMessageType(typ string, fd descriptor.Desc) (string, bool) {
	for _, t := range fd.GetMessageTypes() {
		if t.GetFullyQualifiedName() == typ {
			return fd.GetFullyQualifiedName(), true
		}
	}
	return "", false
}

func findMessageDefLocationFromDependencies(typ string, fd descriptor.Desc) (string, bool) {
	for _, dep := range fd.GetDependencies() {
		for _, t := range dep.GetMessageTypes() {
			if t.GetFullyQualifiedName() == typ {
				return dep.GetFullyQualifiedName(), true
			}
		}
	}
	return "", false
}

func getImports(fd descriptor.Desc, nfd *descriptor.FileDescriptor) ([]string, []descriptor.ImportDesc) {
	imports := []string{}
	// Avoid importing the same package multiple times.
	// Goimports can solve the issue of "import but unused",
	// but it cannot solve problems like "redeclared as imported package name".
	existed := make(map[string]struct{})
	importName2Path := make(map[string]string)

	// Use placeholders to avoid special cases:
	// if the suffix of the "go_package" field defined in the proto file is "proto",
	// the package in "*.trpc.go" will be "proto".
	// However, if other proto files imported by the proto file have a "go_package" suffix of "proto",
	// they can only be numbered starting from "proto1".
	name, path := lang.ExplodeImport(nfd.GoPackage)
	if name == "proto" {
		importName2Path["proto"] = ""
	}
	// Skip the current go_package path to prevent duplicate import.
	existed[path] = struct{}{}
	// The "trpc" name has already been occupied by the trpc-go main library,
	// so it also needs to be skipped.
	importName2Path[trpcName] = ""
	for _, dep := range fd.GetDependencies() {
		pb := dep.GetName()
		pbImport, ok := nfd.Pb2ImportPath[pb]
		if !ok {
			panic(fmt.Errorf("get import path of %s fail", pb))
		}
		_, ok = existed[pbImport]
		if ok { // Prevent duplicate imports.
			continue
		}
		imports = append(imports, pbImport)
		existed[pbImport] = struct{}{}
		importName, importPath := lang.ExplodeImport(pbImport)
		v, ok := importName2Path[importName]
		// If there is no duplication, the importName can be directly used with the suffix.
		if !ok {
			importName2Path[importName] = importPath
			continue
		}
		// If there is a duplication, first check if the importpath is completely the same.
		if importPath == v {
			continue
		}
		// If there is a duplication and the importpath is different, automatic numbering is required.
		// `importName == "proto" && v == ""` indicates the placeholder proto set above.
		var seqno int
		if !((importName == "proto" || importName == trpcName) && v == "") {
			importName2Path[importName+"1"] = v
			delete(importName2Path, importName)
			seqno = 2
		} else {
			seqno = 1
		}
		for i := seqno; ; i++ {
			k := fmt.Sprintf("%s%d", importName, i)
			if _, ok := importName2Path[k]; !ok {
				importName2Path[k] = importPath
				break
			}
		}
	}

	importsX := []descriptor.ImportDesc{}
	for k, v := range importName2Path {
		if (k == "proto" || k == trpcName) && v == "" {
			continue
		}
		desc := descriptor.ImportDesc{
			Name: k,
			Path: v,
		}
		importsX = append(importsX, desc)
	}
	sort.Slice(importsX, func(i, j int) bool {
		return importsX[i].Name <= importsX[j].Name
	})
	return imports, importsX
}

// GetPbPackage retrieves the path where the protobuf files are located.
// When a file option (such as go_package) has a corresponding value in fd.FileOptions,
// the returned pbPackage will be in the form of "trpc.group/trpcprotocol/testapp/testserver".
// Otherwise, pbPackage will be in the form of "trpc.testapp.testserver".
func GetPbPackage(fd *descriptor.FileDescriptor, fileOption string) (string, error) {
	// fd.PackageName usually takes the form of "trpc.testapp.testserver"
	pbPackage := fd.PackageName
	// If fileOption is "go_package",
	// the resulting pbPackage will be in the format "trpc.group/trpcprotocol/testapp/testserver".
	if o := fd.FileOptions[fileOption]; o != nil {
		if v := fd.FileOptions[fileOption].(string); len(v) != 0 {
			pbPackage = v
		}
	}
	return pbPackage, nil
}

// GetPackage combines the package directive and option $lang_package to get a valid package name.
// The package directive is like package trpc.testapp.testserver;.
// The option $lang_package is like option go_package="trpc.group/trpcprotocol/testapp/testserver";.
func GetPackage(fd *descriptor.FileDescriptor, language string) (string, error) {
	// fileOption such as go_package
	fileOption := fmt.Sprintf("%s_package", language)
	pbPackage, err := GetPbPackage(fd, fileOption)
	if err != nil {
		return "", err
	}
	switch fileOption {
	case "go_package":
		pbPackage = lang.TrimRight(";", pbPackage)
	case "cpp_package":
		pbPackage = lang.TrimRight(";", pbPackage)
	default:
		log.Error("unknown FileOption: %s", fileOption)
	}

	return pbPackage, nil
}

// CheckSECVEnabled checks if validation rules are defined in the pb.
func CheckSECVEnabled(nfd *descriptor.FileDescriptor) bool {
	if _, ok := nfd.Pkg2ValidGoPkg["validate"]; ok {
		return ok
	}
	if _, ok := nfd.Pkg2ValidGoPkg["trpc.validate"]; ok {
		return ok
	}
	_, ok := nfd.Pkg2ValidGoPkg["trpc.v2.validate"]
	return ok
}

var errAnnotationNotFound = errors.New("annotation //@alias= not found")

func parseComment(leading, trailing string) (string, error) {
	leadingComment, leadingErr := parseAlias(leading)
	trailingComment, trailingErr := parseAlias(trailing)

	if err := checkCommentErr(leadingComment, leadingErr, trailingComment, trailingErr); err != nil {
		return "", err
	}

	if leadingErr == nil {
		return leadingComment, nil
	}
	return trailingComment, nil
}

func checkCommentErr(leadingComment string, leadingErr error, trailingComment string, trailingErr error) error {
	if leadingErr != nil && trailingErr != nil {
		return errAnnotationNotFound
	}

	if isCommentDiff(leadingComment, leadingErr, trailingComment, trailingErr) {
		return fmt.Errorf("leading and trailing aliases conflict")
	}

	return nil
}

func isCommentDiff(leadingComment string, leadingErr error, trailingComment string, trailingErr error) bool {
	return (leadingErr == nil && trailingErr == nil) && (leadingComment != trailingComment)
}

func parseAlias(comment string) (string, error) {
	const marker = "@alias="
	if !strings.Contains(comment, marker) {
		return "", fmt.Errorf("annotation alias %s not found in raw comment %s", marker, comment)
	}

	const expectSplit = 2
	s := strings.Split(comment, marker)
	if len(s) != expectSplit {
		return "", fmt.Errorf(
			"raw comment %s can be split by annotation alias %s into %d parts, expect %d",
			comment, marker, len(s), expectSplit,
		)
	}

	if !notDoubleCommented(s[0]) {
		return "", fmt.Errorf("candidate alias %s is double commented", s[0])
	}

	alias := strings.TrimSpace(s[1])
	if len(alias) == 0 {
		return "", fmt.Errorf("invalid alias after trim space: %s", comment)
	}

	if idx := strings.IndexAny(alias, " \n"); idx > 0 {
		return alias[:idx], nil
	}
	return alias, nil
}

func notDoubleCommented(prefix string) bool {
	// Example of prefix:
	// Return true:
	//     `// rpc Hello1(HelloReq) returns(HelloRsp){} \n // `
	// Return false:
	//     `// rpc Hello1(HelloReq) returns(HelloRsp){} // `
	s := strings.Split(prefix, "//")
	return len(s) <= 2 || // Only has one comment line.
		strings.Contains(s[len(s)-2], "\n") // Comment line is started on a new line.
}
