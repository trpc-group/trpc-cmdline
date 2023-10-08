package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"

	"trpc.group/trpc-go/trpc-cmdline/bindata/compress"
)

var (
	input  = flag.String("input", "", "read data from input, which could be a regular file or directory")
	output = flag.String("output", "", "write transformed data to named *.go, which could be linked with binary")
	gopkg  = flag.String("gopkg", "gobin", "write transformed data to *.go, whose package is $package")
)

var tpl = `package {{.GoPackage}}
var {{.Variable}} = []uint8{
{{ range $idx, $val := .Data }}{{$val}},{{ end }}
}`

func main() {
	flag.Parse()

	// Validate input and output parameters.
	if len(*input) == 0 || len(*gopkg) == 0 {
		fmt.Println("invalid argument: invalid input")
		os.Exit(1)
	}

	// Read input content.
	buf, err := readFromInputSource(*input)
	if err != nil {
		fmt.Printf("read data error: %v\n", err)
		os.Exit(1)
	}

	// Convert the content to a .go file and write it out.
	inputBaseName := filepath.Base(*input)
	if len(*output) == 0 {
		*output = fmt.Sprintf("%s_bindata.go", inputBaseName)
	}

	outputDir, outputBaseName := filepath.Split(*output)
	tplInstance, err := template.New(outputBaseName).Parse(tpl)
	if err != nil {
		fmt.Printf("parse template error: %v\n", err)
		os.Exit(1)
	}
	_ = os.MkdirAll(outputDir, 0777)

	fout, err := os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("open input error: %v", err)
		os.Exit(1)
	}

	err = tplInstance.Execute(fout, &struct {
		GoPackage string
		Variable  string
		Data      []uint8
	}{
		GoPackage: *gopkg,
		Variable:  strcase.ToCamel(outputBaseName),
		Data:      buf,
	})
	if err != nil {
		panic(fmt.Errorf("template execute error: %v", err))
	}

	fmt.Printf("ok, filedata stored to %s\n", *output)
}

// readFromInputSource reads content from an input source,
// which can be a file or a directory.
// The content will be gzipped then being returned.
func readFromInputSource(inputSource string) (data []byte, err error) {
	_, err = os.Lstat(inputSource)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = compress.Tar(inputSource, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
