VERSION := $(shell cat install/VERSION)
PHONY: fmt bindata_release release submodules image


fmt:
	@gofmt -s -w .
	@goimports -w -local trpc.group .

# If you want to construct bindata and then restore to the initial
# status of submodules, use `make release`
#
# If you only want to construct bindata, use `make bindata_release`,
# but be cautious that the files of submodules has been modified, if
# you want to create a merge request, you need to run `make submodules`
# to restore the status.
release: bindata_release submodules

bindata_release:
	# -git submodule update --init --recursive
	# -rm -rf install/submodules/* && git submodule update --remote
	# -cd install/submodules/trpc-protocol && (ls | grep -v "trpc" | xargs rm -rf) && find . -type f ! -name "*.proto" -exec rm -rf '{}' \; && cd -
	@cat config/version.go | grep -oE "(v[0-9]+.[0-9]+.[0-9]+)" | tr -d '\n' > install/VERSION
	@go get -u trpc.group/trpc-go/trpc-cmdline/bindata
	@# in case "go get -u" doesn't work in higher version of Go.
	@go install trpc.group/trpc-go/trpc-cmdline/bindata
	@bindata -input=install -output=gobin/assets.go -gopkg=gobin
	# The dependency on strcase should be forced to v0.2.0 for compatibility.
	@go get -u github.com/iancoleman/strcase@v0.2.0
	@go mod tidy

submodules:
	# -rm -rf install/submodules/* && git submodule update --remote

image:
	@mkdir -p bin
	@GOOS=linux go build -o bin/ trpc/trpc.go
	@docker build -t trpc-cmdline:${VERSION} .
