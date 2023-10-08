# Usage:
#   1. Create stub files: make stub
#   2. Clean all the generated files: make clean

.PHONY: stub clean

stub:
	go mod init trpc.group/examples/helloworld
	trpc create -p common/common.proto -o common --rpconly --usebasename=false 
	trpc create -p service1/service1.proto -o service1 --rpconly --usebasename=false --nogomod
	trpc create -p service2/service2.proto -o service2 --rpconly --usebasename=false --nogomod

clean:
	find . -name "*.go" -type f -delete
	find . -name "*.mod" -type f -delete
	find . -name "*.sum" -type f -delete
