all:
	#trpc create --protofile=greeter.proto --swagger
	trpc apidocs -p helloworld.swagger.proto --swagger

.PHONY: clean

clean:
	rm -rf apidocs.swagger.json

