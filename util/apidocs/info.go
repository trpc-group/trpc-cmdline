package apidocs

import (
	"fmt"
	"path/filepath"
	"strings"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
)

// InfoStruct defines the structure of the documentation description information contained in the apidocs header.
type InfoStruct struct {
	Title       string `json:"title"`                 // Title of the doc.
	Description string `json:"description,omitempty"` // Description of the doc.
	Version     string `json:"version,omitempty"`     // Version of the doc.
}

// NewInfo inits Info instance.
func NewInfo(fd *descriptor.FileDescriptor) (InfoStruct, error) {
	filePath, err := filepath.Abs(fd.FilePath)
	if err != nil {
		return InfoStruct{}, err
	}
	_, fileName := filepath.Split(filePath)
	title := strings.ReplaceAll(fileName, ".proto", "")
	infoMap := InfoStruct{
		Title:       title,
		Description: fmt.Sprintf("The api document of %s", fileName),
		Version:     "2.0",
	}
	return infoMap, nil
}
