all:
	@echo "------------- [trpc create] --------------"
	trpc create --protofile=helloworld.proto -protocol=trpc

.PHONY: clean

clean:
	rm -rf greeter

compile:
	@echo "------------- [compile trpc] -------------"
	cd ../../../trpc-cmdline/
	make clean && make && make install
	@cd -
