package swagger

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
	"trpc.group/trpc-go/trpc-cmdline/params"
	"trpc.group/trpc-go/trpc-cmdline/util/apidocs"
)

func TestGenSwagger(t *testing.T) {
	type args struct {
		fd     *descriptor.FileDescriptor
		option *params.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		newErr  error
	}{
		{
			name: "case1: new err",
			args: args{
				fd:     &descriptor.FileDescriptor{},
				option: &params.Option{},
			},
			wantErr: true,
			newErr:  fmt.Errorf("err"),
		},
		{
			name: "case1: without err",
			args: args{
				fd:     &descriptor.FileDescriptor{},
				option: &params.Option{},
			},
			wantErr: false,
			newErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := gomonkey.ApplyFunc(
				apidocs.NewSwagger,
				func(fd *descriptor.FileDescriptor, option *params.Option) (*apidocs.SwaggerJSON, error) {
					return &apidocs.SwaggerJSON{}, tt.newErr
				},
			).ApplyFunc(
				apidocs.WriteJSON,
				func(file string, data interface{}) error {
					return nil
				},
			)

			defer p.Reset()

			if err := GenSwagger(tt.args.fd, tt.args.option); (err != nil) != tt.wantErr {
				t.Errorf("GenSwagger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
