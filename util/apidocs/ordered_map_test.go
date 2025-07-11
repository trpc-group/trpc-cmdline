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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedMarshalJSON(t *testing.T) {
	type args struct {
		element interface{}
		rank    map[string]int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "marshal",
			args: args{
				element: map[string]string{
					"1": "value-1",
					"2": "value-2",
					"3": "value-3",
				},
				rank: map[string]int{
					"1": 1,
					"2": 2,
					"3": 3,
				},
			},
			want: []byte(`{"1":"value-1","2":"value-2","3":"value-3"}`),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
		{
			name: "marshal in another order",
			args: args{
				element: map[string]string{
					"1": "value-1",
					"2": "value-2",
					"3": "value-3",
				},
				rank: map[string]int{
					"2": 1,
					"1": 2,
					"3": 3,
				},
			},
			want: []byte(`{"2":"value-2","1":"value-1","3":"value-3"}`),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
		{
			name: "marshal in standard order",
			args: args{
				element: map[string]string{
					"2": "value-2",
					"1": "value-1",
					"3": "value-3",
				},
				rank: make(map[string]int),
			},
			want: []byte(`{"1":"value-1","2":"value-2","3":"value-3"}`),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrderedMarshalJSON(tt.args.element, tt.args.rank)
			tt.wantErr(t, err, fmt.Sprintf("OrderedMarshalJSON(%v, %v)", tt.args.element, tt.args.rank))
			require.Equalf(t, tt.want, got, "OrderedMarshalJSON(%v, %v)", tt.args.element, tt.args.rank)
		})
	}
}

func TestOrderedUnmarshalJSON(t *testing.T) {
	type args struct {
		b       []byte
		element interface{}
		rank    interface{}
	}
	var emptyRank map[string]int
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "unmarshal",
			args: args{
				b:       []byte(`{"2":"value-2","1":"value-1","3":"value-3"}`),
				element: &map[string]string{},
				rank:    &map[string]int{},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
		{
			name: "unmarshal_with_nil",
			args: args{
				b:       []byte(`{"2":"value-2","1":"value-1","3":"value-3"}`),
				element: &map[string]string{},
				rank:    &emptyRank,
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
		{
			name: "unmarshal_rank",
			args: args{
				b: []byte(`{
					 "domain": {
					  "title": "domain",
					  "$ref": "#/components/schemas/helloworld.Domain.aa6718f0a7c001e99386d62d6a0da155"
					 },
					 "url": {
					  "title": "url",
					  "type": "string"
					 }
				}`),
				element: &map[string]interface{}{},
				rank:    &emptyRank,
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, OrderedUnmarshalJSON(tt.args.b, tt.args.element, tt.args.rank),
				fmt.Sprintf("OrderedUnmarshalJSON(%v, %v, %v)", tt.args.b, tt.args.element, tt.args.rank))
		})
	}
}
