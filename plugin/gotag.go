// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package plugin

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	trpc "trpc.group/trpc/trpc-protocol/pb/go/trpc/proto"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	tparser "trpc.group/trpc-go/trpc-cmdline/parser"
	"trpc.group/trpc-go/trpc-cmdline/util/fs"
	"trpc.group/trpc-go/trpc-cmdline/util/log"
)

var (
	regexpInject          = regexp.MustCompile("`.+`$")
	regexpTags            = regexp.MustCompile(`[\w_]+:"[^"]+"`)
	regexpProtobufTagName = regexp.MustCompile(`protobuf:"[\w_\-,]+name=([\w_\-]+)`)
)

// textArea records tag text position and tag info in *.pb.go file
type textArea struct {
	StartPos   int
	EndPos     int
	CurrentTag string
	NewTag     string
}

// GoTag generates go tag by proto field options
type GoTag struct {
}

// Name return plugin's name
func (p *GoTag) Name() string {
	return "gotag"
}

// Check only run when `--lang=go && --go_tag=true`
func (p *GoTag) Check(fd *descriptor.FileDescriptor, opt *params.Option) bool {
	if opt.Language == "go" && opt.Gotag {
		return true
	}
	return false
}

// Run exec go tag plugin
func (p *GoTag) Run(fd *descriptor.FileDescriptor, opt *params.Option) error {
	tags := optTagsFromProto(fd.FD)
	if len(tags) == 0 {
		return nil
	}

	outputdir := opt.OutputDir
	pbfile := ""
	pbname := fs.BaseNameWithoutExt(fd.FilePath) + ".pb.go"

	if opt.RPCOnly {
		pbfile = filepath.Join(outputdir, pbname)
	} else {
		importPath, err := tparser.GetPbPackage(fd, "go_package")
		if err != nil {
			return err
		}
		pbfile = filepath.Join(outputdir, "stub", importPath, pbname)
	}

	return p.replaceTags(pbfile, tags)
}

func (p *GoTag) replaceTags(pbfile string, tags map[string]string) error {
	_, err := os.Lstat(pbfile)
	if err != nil {
		return err
	}
	areas, err := tagAreasFromPBFile(pbfile, tags)
	if err != nil {
		return err
	}
	if err = injectTagsToPBFile(pbfile, areas); err != nil {
		return err
	}
	return nil
}

// optTagsFromProto parses field go tag option from proto file and maps it as a kv map
// map structure should be like `messageName_fieldName`
func optTagsFromProto(fd descriptor.Desc) map[string]string {
	tagmap := make(map[string]string)
	var scanNestedMsgFunc func(*desc.MessageDescriptor, string)
	scanNestedMsgFunc = func(m *desc.MessageDescriptor, prefix string) {
		for _, mm := range m.GetNestedMessageTypes() {
			p := fmtgotagkey(prefix, m.GetName())
			scanNestedMsgFunc(mm, p)
		}
		for _, field := range m.GetFields() {
			tags := getGoTag(field.GetFieldOptions())
			if tags == "" {
				continue
			}
			key := fmtgotagkey(prefix, m.GetName(), field.GetName())
			tagmap[key] = tags
		}
	}
	for _, msg := range fd.GetMessageTypes() {
		messageDescriptor, ok := msg.(*descriptor.ProtoMessageDescriptor)
		if !ok {
			continue
		}
		md := messageDescriptor.MD
		scanNestedMsgFunc(md, "")
	}
	return tagmap
}

func getGoTag(opts *descriptorpb.FieldOptions) string {
	if proto.HasExtension(opts, trpc.E_GoTag) {
		return proto.GetExtension(opts, trpc.E_GoTag).(string)
	}
	return ""
}

// fmtgotagkey generates the key for `protoTags` to join the struct name and
// field name by `_`, nested message names would be joined too
func fmtgotagkey(s ...string) string {
	for k, v := range s {
		if v == "" {
			s = append(s[:k], s[k+1:]...)
		}
	}
	return strcase.ToCamel(strings.Join(s, "_"))
}

// tagAreasFromPBFile parses *.pb.go and records tag positions which need to be replaced
func tagAreasFromPBFile(fp string, newtags map[string]string) (areas []textArea, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fp, nil, parser.ParseComments)
	if err != nil {
		return
	}
	for _, decl := range f.Decls {
		// check if is generic declaration
		typeSpec := genTypeSpec(decl)
		// skip if can't get type spec
		if typeSpec == nil {
			continue
		}
		// not a struct, skip
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		areas = append(areas, genAreas(structDecl, typeSpec, newtags)...)
	}
	return
}

func genAreas(structDecl *ast.StructType, typeSpec *ast.TypeSpec, newtags map[string]string) []textArea {
	var areas []textArea
	for _, field := range structDecl.Fields.List {
		if field.Tag == nil {
			continue
		}
		fieldname := protobufTagName(field.Tag.Value)
		if fieldname == "" {
			continue
		}
		structname := typeSpec.Name.String()
		// key = structName_fieldName
		key := fmtgotagkey(structname, fieldname)
		newtag, ok := newtags[key]
		if !ok {
			continue
		}
		currentTag := field.Tag.Value
		areas = append(areas, textArea{
			StartPos:   int(field.Pos()),
			EndPos:     int(field.End()),
			CurrentTag: currentTag[1 : len(currentTag)-1],
			NewTag:     newtag,
		})
	}
	return areas
}

func genTypeSpec(decl ast.Decl) *ast.TypeSpec {
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok {
		return nil
	}
	var typeSpec *ast.TypeSpec
	for _, spec := range genDecl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			typeSpec = ts
			break
		}
	}
	return typeSpec
}

func protobufTagName(tag string) string {
	matches := regexpProtobufTagName.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// injectTagsToPBFile replaces tags and rewrites the *.pb.go file
func injectTagsToPBFile(fp string, areas []textArea) (err error) {
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	contents, err := io.ReadAll(f)
	if err != nil {
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	return writeTagsToFile(fp, areas, contents, err)
}

func writeTagsToFile(fp string, areas []textArea, contents []byte, err error) error {
	// inject custom tags from tail of file first to preserve order
	for i := range areas {
		area := areas[len(areas)-i-1]
		log.Debug("inject custom tag %q to expression %q",
			area.NewTag, string(contents[area.StartPos-1:area.EndPos-1]))
		contents = injectGoTag(contents, area)
	}
	if err = os.WriteFile(fp, contents, 0644); err != nil {
		return err
	}
	if len(areas) > 0 {
		log.Debug("file %q is injected with custom tags", fp)
	}
	return nil
}

func injectGoTag(contents []byte, area textArea) (injected []byte) {
	expr := make([]byte, area.EndPos-area.StartPos)
	copy(expr, contents[area.StartPos-1:area.EndPos-1])
	cti := newGoTagItems(area.CurrentTag)
	iti := newGoTagItems(area.NewTag)
	ti := cti.override(iti)
	expr = regexpInject.ReplaceAll(expr, []byte(fmt.Sprintf("`%s`", ti.format())))
	injected = append(injected, contents[:area.StartPos-1]...)
	injected = append(injected, expr...)
	injected = append(injected, contents[area.EndPos-1:]...)
	return
}

type goTagItem struct {
	key   string
	value string
}

type goTagItems []goTagItem

func (ti goTagItems) format() string {
	tags := []string{}
	for _, item := range ti {
		tags = append(tags, fmt.Sprintf(`%s:%s`, item.key, item.value))
	}
	return strings.Join(tags, " ")
}

func (ti goTagItems) override(nti goTagItems) goTagItems {
	overridden := []goTagItem{}
	for i := range ti {
		var dup = -1
		for j := range nti {
			if ti[i].key == nti[j].key {
				dup = j
				break
			}
		}
		if dup == -1 {
			overridden = append(overridden, ti[i])
		} else {
			overridden = append(overridden, nti[dup])
			nti = append(nti[:dup], nti[dup+1:]...)
		}
	}
	return append(overridden, nti...)
}

func newGoTagItems(tag string) goTagItems {
	var items goTagItems
	split := regexpTags.FindAllString(tag, -1)
	for _, t := range split {
		sepPos := strings.Index(t, ":")
		items = append(items, goTagItem{
			key:   t[:sepPos],
			value: t[sepPos+1:],
		})
	}
	return items
}
