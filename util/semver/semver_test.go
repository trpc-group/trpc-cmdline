// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package semver

import "testing"

func TestNewerThan(t *testing.T) {
	type args struct {
		v1 string
		v2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"1.0.0 vs 1.0.0", args{"1.0.0", "1.0.0"}, false},
		{"1.0.0 vs 1.0.1", args{"1.0.0", "1.0.1"}, false},
		{"1.0.0 vs 1.1.0", args{"1.0.0", "1.1.0"}, false},
		{"1.0.0 vs 1.1.1", args{"1.0.0", "1.1.1"}, false},
		{"2.0.0 vs 1.0.0", args{"2.0.0", "1.0.0"}, true},
		{"2.0.0 vs 1.0.1", args{"2.0.0", "1.0.1"}, true},
		{"2.0.0 vs 1.1.0", args{"2.0.0", "1.1.0"}, true},
		{"2.0.0 vs 1.1.1", args{"2.0.0", "1.1.1"}, true},
		{"2.5.0 vs 2.6.0", args{"2.5.0", "2.6.0"}, false},
		{"2.5.0 vs 2.6.1", args{"2.5.0", "2.6.1"}, false},
		{"2.5.1 vs 2.5.1", args{"2.5.1", "2.5.1"}, false},
		{"2.6.0 vs 2.6.0", args{"2.6.0", "2.6.0"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewerThan(tt.args.v1, tt.args.v2); got != tt.want {
				t.Errorf("NewerThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersions(t *testing.T) {
	tests := []struct {
		name         string
		ver          string
		wantMajor    int
		wantMinor    int
		wantRevision int
	}{
		{"v1.2.3-dev", "v1.2.3-dev", 1, 2, 3},
		{"1.2.3-dev", "1.2.3-dev", 1, 2, 3},
		{"1.2.3", "1.2.3", 1, 2, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, gotRevision := Versions(tt.ver)
			if gotMajor != tt.wantMajor {
				t.Errorf("Versions() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotMinor != tt.wantMinor {
				t.Errorf("Versions() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
			}
			if gotRevision != tt.wantRevision {
				t.Errorf("Versions() gotRevision = %v, want %v", gotRevision, tt.wantRevision)
			}
		})
	}
}
