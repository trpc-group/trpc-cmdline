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
	"crypto/md5"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
)

// Paths is the set of Path.
type Paths struct {
	Elements map[string]Methods
	Rank     map[string]int
}

// Put puts the element into the ordered map and records the element's ranking order in "Rank".
func (paths *Paths) Put(key string, value Methods) {
	paths.Elements[key] = value

	if paths.Rank != nil {
		if _, ok := paths.Rank[key]; !ok {
			paths.Rank[key] = len(paths.Elements)
		}
	}
}

// UnmarshalJSON deserializes JSON data
func (paths *Paths) UnmarshalJSON(b []byte) error {
	return OrderedUnmarshalJSON(b, &paths.Elements, &paths.Rank)
}

// MarshalJSON serializes to JSON
func (paths Paths) MarshalJSON() ([]byte, error) {
	return OrderedMarshalJSON(paths.Elements, paths.Rank)
}

// NewPaths inits Path.
func NewPaths(fd *descriptor.FileDescriptor, option *params.Option, defs *Definitions) (Paths, error) {
	paths := Paths{
		Elements: make(map[string]Methods),
	}

	if option.OrderByPBName {
		paths.Rank = make(map[string]int)
	}

	var err error
	for _, service := range fd.Services {
		// service.RPC contains the original RPC method and its restful bindings.
		for _, rpc := range service.RPC {
			args := methodArgs{
				RPC:  rpc,
				Defs: defs,
				Opt:  option,
			}
			if !option.AliasOn || option.KeepOrigRPCName ||
				// If alias is set to true and keep-orig-rpcname is set to false, but the RPC method
				// does not have an alias, the original RPC information should still be displayed.
				len(service.MethodRPCx[rpc.Name]) == 0 {
				args.Tags = []string{strings.ToLower(service.Name) + "." + "trpc"}
				paths.addRPCMethod(args)
			}
			args.Tags = []string{strings.ToLower(service.Name) + "." + "restful"}
			if e := paths.addRestfulMethod(args); e != nil {
				err = multierror.Append(err, e).ErrorOrNil()
			}
		}
		// service.RPCx only contains the alias RPC method which excludes any restful bindings.
		for _, rpc := range service.RPCx {
			args := methodArgs{
				RPC:  rpc,
				Defs: defs,
				Opt:  option,
			}
			args.Tags = []string{strings.ToLower(service.Name) + "." + "trpc"}
			paths.addRPCMethod(args)
		}
	}

	paths.cleanOperationID()
	return paths, err
}

// methodArgs adds a method to Paths with the given method arguments.
type methodArgs struct {
	RPC  *descriptor.RPCDescriptor
	Defs *Definitions
	Tags []string
	Opt  *params.Option
}

func (args methodArgs) summary() string {
	// Get the description of each rpc method defined in front (if any).
	summary := args.RPC.LeadingComments
	// Verify the names of the rpc methods collected in the "option" previously.
	if len(args.RPC.SwaggerInfo.Title) != 0 {
		summary = args.RPC.SwaggerInfo.Title
	}
	return trimExtraneous(summary)
}

func trimExtraneous(input string) string {
	const marker = "@alias="
	s := strings.Split(input, marker)
	// Remove the alias if had any.
	output := s[0]
	// Remove comment slashes.
	output = strings.ReplaceAll(output, "\n//", " ")
	output = strings.ReplaceAll(output, "//", " ")
	return strings.Trim(output, " \"'")
}

func (args methodArgs) method() *MethodStruct {
	return &MethodStruct{
		Summary:     args.summary(),
		OperationID: args.RPC.Name,
		Responses:   args.Defs.getMediaStruct(args.RPC.ResponseType),
		Tags:        args.Tags,
		Description: args.RPC.SwaggerInfo.Description,
	}
}

func (args methodArgs) rpcParams() []*ParametersStruct {
	queryParams := args.Defs.getQueryParameters(args.RPC.RequestType)
	if args.Opt.SwaggerOptJSONParam {
		queryParams = args.Defs.getBodyParameters(args.RPC.RequestType)
	}

	args.fillDescriptorToParams(queryParams)
	return queryParams
}

func (args methodArgs) restfulParams(api descriptor.RESTfulAPIContent) ([]*ParametersStruct, error) {
	pathParams := newPathParams(api.PathTmpl)

	names := pathParams.getNames()
	reqType := args.RPC.RequestType
	if len(names) > 0 {
		suffix := fmt.Sprintf("%x", md5.Sum([]byte(api.PathTmpl)))
		args.Defs.filterFields(reqType, suffix, names)
		reqType += "." + suffix
	}

	params := pathParams.getParameters()
	method := strings.ToLower(api.Method)

	if api.RequestBody == "" && (method == "get" || method == "delete") {
		params = append(params, args.Defs.getQueryParameters(reqType)...)
	}

	if api.RequestBody == "*" {
		params = append(params, args.Defs.getBodyParameters(reqType)...)
	}

	if api.RequestBody != "" && api.RequestBody != "*" {
		param, err := args.Defs.getBodyParameter(reqType, api.RequestBody)
		if err != nil {
			return nil, fmt.Errorf("generate restful parameter error: %w", err)
		}
		params = append(params, param)
	}

	args.fillDescriptorToParams(params)

	return params, nil
}

func (args methodArgs) fillDescriptorToParams(params []*ParametersStruct) {
	for _, param := range params {
		spd, ok := args.RPC.SwaggerInfo.Params[param.Name]
		if ok {
			param.Required = spd.Required
			if spd.Default != "" {
				param.Default = spd.Default
			}
		}
	}
}

// GetPathsX converts to  openapi v3 structure.
func (paths Paths) GetPathsX() PathsX {
	pathsX := PathsX{Elements: make(map[string]MethodsX)}
	pathsX.Rank = paths.Rank
	paths.orderedEach(func(path string, methods Methods) {
		pathsX.Elements[path] = methods.GetMethodsX()
	})
	return pathsX
}

func (paths Paths) addRPCMethod(args methodArgs) {
	method := args.method()
	method.Parameters = args.rpcParams()

	mx := Methods{Elements: make(map[string]*MethodStruct)}
	if paths.Rank != nil {
		mx.Rank = make(map[string]int)
	}
	mx.Put(args.RPC.SwaggerInfo.Method, method)
	paths.Put(args.RPC.FullyQualifiedCmd, mx)
}

func (paths Paths) addRestfulMethod(args methodArgs) error {
	var err error
	for _, api := range args.RPC.RESTfulAPIInfo.ContentList {
		// Filter out the existing paths
		path := api.PathTmpl
		pathParams := newPathParams(api.PathTmpl)
		for _, param := range pathParams {
			if param.value != "" {
				path = strings.Replace(path, param.origin, param.value, -1)
			}
		}

		params, e := args.restfulParams(*api)
		if e != nil {
			err = multierror.Append(err, e).ErrorOrNil()
		}
		method := args.method()
		method.Parameters = params

		mx, ok := paths.Elements[path]
		if !ok {
			mx = Methods{
				Elements: make(map[string]*MethodStruct),
			}
			if paths.Rank != nil {
				mx.Rank = make(map[string]int)
			}
		}

		mx.Put(strings.ToLower(api.Method), method)
		paths.Put(path, mx)
	}
	return err
}

// orderedEach sort each element
func (paths *Paths) orderedEach(f func(path string, methods Methods)) {
	if paths == nil {
		return
	}

	var keys []string
	for k := range paths.Elements {
		keys = append(keys, k)
	}

	if paths.Rank != nil {
		sort.Slice(keys, func(i, j int) bool {
			return paths.Rank[keys[i]] < paths.Rank[keys[j]]
		})
	} else {
		sort.Strings(keys)
	}

	for _, k := range keys {
		f(k, paths.Elements[k])
	}
}

// cleanOperationID adds a suffix number to the end of the OperationID to avoid duplication.
// The reason for using "order each" is that the loop over the map is random,
// which leads to unstable results and cannot be tested stably.
func (paths Paths) cleanOperationID() {
	operationIDSet := make(map[string]int)
	paths.orderedEach(func(path string, methods Methods) {
		methods.orderedEach(func(k string, method *MethodStruct) {
			operationIDSet[method.OperationID]++
			if operationIDSet[method.OperationID] > 1 {
				method.OperationID = fmt.Sprintf("%s%d",
					method.OperationID, operationIDSet[method.OperationID])
			}
		})
	})
}

type pathParam struct {
	name   string
	value  string
	origin string
}

type pathParams []pathParam

func (params pathParams) getParameters() []*ParametersStruct {
	parameters := make([]*ParametersStruct, 0)
	for _, param := range params {
		if param.value != "" {
			continue
		}

		parameters = append(parameters, &ParametersStruct{
			Name:     param.name,
			Default:  "",
			Required: true,
			Type:     "string",
			In:       "path",
		})
	}

	return parameters
}

func (params pathParams) getNames() []string {
	var names []string
	for _, param := range params {
		names = append(names, param.name)
	}
	return names
}

var pathParamsRE = regexp.MustCompile("{.*?}")

func newPathParams(path string) pathParams {
	var params []pathParam
	values := pathParamsRE.FindAllString(path, -1)

	for _, v := range values {
		pos := strings.Index(v, "=")
		if pos == -1 {
			params = append(params, pathParam{
				origin: v,
				name:   v[1 : len(v)-1],
			})
			continue
		}
		params = append(params, pathParam{
			origin: v,
			name:   v[1:pos],
			value:  v[pos+1 : len(v)-1],
		})
	}
	return params
}
