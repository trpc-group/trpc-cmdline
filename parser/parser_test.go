package parser

import (
	"testing"

	annotations "trpc.group/trpc/trpc-protocol/pb/go/trpc/api"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
)

func Test_parseRPCAlias(t *testing.T) {
	tests := []struct {
		name        string
		comment     string
		wantComment string
		wantErr     bool
	}{
		{
			name:        "1-annotation_ok",
			comment:     "//@alias=/api/hello",
			wantComment: "/api/hello",
			wantErr:     false,
		}, {
			name:        "2-annotation_ok (with comments)",
			comment:     "//@alias=/api/hello 其他描述",
			wantComment: "/api/hello",
			wantErr:     false,
		}, {
			name:        "3-annotation_ok (with delimiter)",
			comment:     "//@alias= /api/hello 	其他描述",
			wantComment: "/api/hello",
			wantErr:     false,
		}, {
			name:        "4-annotation_err",
			comment:     "//@alia= /api/hello 	其他描述",
			wantComment: "",
			wantErr:     true,
		}, {
			name:        "5-annotation_err (must occur once)",
			comment:     "//@alias=/api/hello @alias=/api/hello",
			wantComment: "",
			wantErr:     true,
		}, {
			name:        "6-annotation_err (empty value)",
			comment:     "//@alias=   ",
			wantComment: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComment, err := parseAlias(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAnnotatedComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotComment != tt.wantComment {
				t.Errorf("getAnnotatedComment() gotComment = %v, want %v", gotComment, tt.wantComment)
			}
		})
	}
}

func Test_parseRPCComment(t *testing.T) {
	tests := []struct {
		name        string
		rpc         *descriptor.RPCDescriptor
		wantComment string
		wantErr     bool
	}{
		{
			name: "1-annotation_not_found",
			rpc: &descriptor.RPCDescriptor{
				LeadingComments:  "//@alia=/api/hello",
				TrailingComments: "//@alia=/api/hello",
			},
			wantComment: "",
			wantErr:     true,
		}, {
			name: "2-annotation_conflict",
			rpc: &descriptor.RPCDescriptor{
				LeadingComments:  "//@alias=/api/hello1",
				TrailingComments: "//@alias=/api/hello2",
			},
			wantComment: "",
			wantErr:     true,
		}, {
			name: "3-select_valid_annotation (leading)",
			rpc: &descriptor.RPCDescriptor{
				LeadingComments:  "//@alias=/api/hello",
				TrailingComments: "//@alia=/api/hello",
			},
			wantComment: "/api/hello",
			wantErr:     false,
		}, {
			name: "4-select_valid_annotation (trailing)",
			rpc: &descriptor.RPCDescriptor{
				LeadingComments:  "//@alia=/api/hello",
				TrailingComments: "//@alias=/api/hello",
			},
			wantComment: "/api/hello",
			wantErr:     false,
		}, {
			name: "5-select_valid_annotation (same)",
			rpc: &descriptor.RPCDescriptor{
				LeadingComments:  "//@alias=/api/hello",
				TrailingComments: "//@alias=/api/hello",
			},
			wantComment: "/api/hello",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComment, err := parseComment(tt.rpc.LeadingComments, tt.rpc.TrailingComments)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAnnotationValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotComment != tt.wantComment {
				t.Errorf("getAnnotationValue() gotComment = %v, want %v", gotComment, tt.wantComment)
			}
		})
	}
}

func Test_fillMethodRESTfulInfo(t *testing.T) {
	tests := []struct {
		name     string
		httpRule *annotations.HttpRule
		wantErr  bool
	}{
		{
			name: "case-unknown_RESTful_method",
			httpRule: &annotations.HttpRule{
				Selector:           "",
				Pattern:            nil,
				Body:               "",
				ResponseBody:       "",
				AdditionalBindings: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseRestContents(tt.httpRule); (err != nil) != tt.wantErr {
				t.Errorf("fillRESTfulAPIContentMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
