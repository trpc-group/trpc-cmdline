Moved from: https://trpc.group/gitcode/bindata

# Introduction

`bindata` is used for converting file or directory into a go file, whose package name can be specified.

The original file data is stored in a exported variable, so it can be referenced in your code.

# Installation

```
go get -u trpc.group/trpc-go/trpc-cmdline/bindata
```
# Help

```bash
$ bindata -h

Usage of bindata:
  -gopkg string
    	write transformed data to *.go, whose package is $package (default "gobin")
  -input string
    	read data from input, which could be a regular file or directory
  -output string
    	write transformed data to named *.go, which could be linked with binary
```
