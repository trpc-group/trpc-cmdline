// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the  Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package lang

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/require"

	"trpc.group/trpc-go/trpc-cmdline/descriptor"
)

func TestPBSimplifyGoType(t *testing.T) {
	t.Run("case without dot", func(t *testing.T) {
		require.Panics(t, func() {
			PBSimplifyGoType("pkgname::typename", "pkgname")
		})
	})

	t.Run("case pkg == goPackageName", func(t *testing.T) {
		p := PBSimplifyGoType("helloworld.HelloRequest", "helloworld")
		require.Equal(t, "HelloRequest", p)
	})

	t.Run("case pkg != goPackageName", func(t *testing.T) {
		p := PBSimplifyGoType("pkgname.typename", "pkgname")
		require.Equal(t, "typename", p)
	})
}

func TestPBGoType(t *testing.T) {
	t.Run("case without dot", func(t *testing.T) {
		require.Panics(t, func() {
			PBGoType("typeWithoutDot")
		})
	})

	t.Run("case succ", func(t *testing.T) {
		p := PBGoType("prefix/a.b.c.hello")
		require.Equal(t, "prefixa_b_c.Hello", p)
	})
}

func TestPBGoPackage(t *testing.T) {
	require.Equal(t, "a_b_c", PBGoPackage("a.b.c"))
	require.Equal(t, "a_b_c", PBGoPackage("prefix/a.b.c"))
	require.Equal(t, "a_b_c", PBGoPackage("prefix/a-b-c"))
	require.Equal(t, "a_b_c", PBGoPackage("a-b-c"))
	require.Equal(t, "a_b_c", PBGoPackage("a.b-c"))
}

func TestGoExport(t *testing.T) {
	t.Run("case without dot", func(t *testing.T) {
		typ := GoExport("test")
		require.Equal(t, "Test", typ)
	})

	t.Run("case with dot", func(t *testing.T) {
		typ := GoExport("test.test")
		require.Equal(t, "test.Test", typ)
	})
}

func TestSplitList(t *testing.T) {
	vals := SplitList(",", "a,b,c")
	require.Equal(t, []string{"a", "b", "c"}, vals)
}

func TestTrimRight(t *testing.T) {
	t.Run("case no split", func(t *testing.T) {
		s := TrimRight(",", "test")
		require.Equal(t, "test", s)
	})

	t.Run("case with split", func(t *testing.T) {
		s := TrimRight(".", "test.go")
		require.Equal(t, "test", s)
	})
}

func TestTrimLeft(t *testing.T) {
	s := TrimLeft(".", "test.go")
	require.Equal(t, "go", s)
}

func TestTitle(t *testing.T) {
	s := Title("test")
	require.Equal(t, "Test", s)
}

func TestUnTitle(t *testing.T) {
	s := UnTitle("TEst")
	require.Equal(t, "tEst", s)
}

func TestGoFullyQualifiedType(t *testing.T) {
	t.Run("case invalid go type", func(t *testing.T) {
		require.Panics(t, func() {
			GoFullyQualifiedType("test/test", &descriptor.FileDescriptor{})
		})
	})

	t.Run("case get pb empty", func(t *testing.T) {
		fullTyp := "trpc.group/sample/helloworld.Request"
		typ := GoFullyQualifiedType(fullTyp, &descriptor.FileDescriptor{})
		require.Equal(t, fullTyp, typ)
	})
}

func TestPBValidGoPackage(t *testing.T) {
	t.Run("case without /", func(t *testing.T) {
		pkg := PBValidGoPackage("test.test")
		require.Equal(t, "test_test", pkg)
	})

	t.Run("case with /", func(t *testing.T) {
		pkg := PBValidGoPackage("test/test.test")
		require.Equal(t, "test_test", pkg)
	})
}

func TestLast(t *testing.T) {
	ret := Last([]string{"a", "b", "c"})
	require.Equal(t, "c", ret)
}

func TestHasPrefix(t *testing.T) {
	has := HasPrefix("prefix", "prefix/test")
	require.True(t, has)
}

func TestHasSuffix(t *testing.T) {
	has := HasSuffix("suffix", "test/suffix")
	require.True(t, has)
}

func TestAdd(t *testing.T) {
	ret := Add(1, 2)
	require.Equal(t, 3, ret)
}

func TestLoadGoMod(t *testing.T) {
	t.Run("case panic", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "", errors.New("fake error")
		})
		defer p.Reset()
		require.Panics(t, func() {
			LoadGoMod()
		})
	})

	t.Run("case Lstat error", func(t *testing.T) {
		p := gomonkey.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd/", nil
		})
		p.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("fake error")
		})
		defer p.Reset()
		mod, err := LoadGoMod()
		require.Empty(t, mod)
		require.NotNil(t, err)
	})

	t.Run("case open error", func(t *testing.T) {
		p := gomonkey.NewPatches()
		defer p.Reset()
		p.ApplyFunc(os.Getwd, func() (dir string, err error) {
			return "wd/", nil
		})
		p.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		p.ApplyFunc(os.Open, func(name string) (*os.File, error) {
			return nil, errors.New("fake error")
		})
		mod, err := LoadGoMod()
		require.Empty(t, mod)
		require.NotNil(t, err)
	})
}

func TestCheckSECVTpl(t *testing.T) {
	t.Run("case true", func(t *testing.T) {
		ret := CheckSECVTpl(map[string]string{
			"validate": "exist",
		})
		require.True(t, ret)
	})

	t.Run("case false", func(t *testing.T) {
		ret := CheckSECVTpl(make(map[string]string))
		require.False(t, ret)
	})
}

func TestCamelcase(t *testing.T) {
	t.Run("case empty", func(t *testing.T) {
		ret := Camelcase("")
		require.Empty(t, ret)
	})

	t.Run("case one word", func(t *testing.T) {
		ret := Camelcase("TEST")
		require.Equal(t, "TEST", ret)
		ret = Camelcase("test")
		require.Equal(t, "Test", ret)
	})

	t.Run("case multi words", func(t *testing.T) {
		ret := Camelcase("test_test")
		require.Equal(t, "TestTest", ret)
		ret = Camelcase("test_ABC_ABC")
		require.Equal(t, "Test_ABC_ABC", ret)
	})
}

func Test_isAllUpper(t *testing.T) {
	t.Run("case false", func(t *testing.T) {
		ret := isAllUpper("Abc")
		require.False(t, ret)
	})

	t.Run("case true", func(t *testing.T) {
		ret := isAllUpper("ABC")
		require.True(t, ret)
	})
}
