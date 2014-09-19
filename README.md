# Goprotoc 

Goprotoc is a new experimental go protobuf compiler.  It is a fork of GoGoProtobuf http://code.google.com/p/gogoprotobuf, which extends GoProtobuf http://code.google.com/p/goprotobuf

The main goal of this project is to reduce memory pressure/gc pauses and to provide useful API for accessing/mutating message fields.

### Current state

Goprotoc should be considered in early alpha; it "works" that in can successfully compile proto files
on a few number of testcases but is not used in any production environment.

### Contributing

Goprotoc welcomes any kind of contribution; please send any pull request for feature, tests or fixes.
> tl;dr: You will need to sign the [Dropbox CLA](https://opensource.dropbox.com/cla/) and run the tests.

### Roadmap

##### v0.1: [released 9/19/2014]
- Remove pointers from messages, add special fields to check is set
- Generate full API functions to set/get/clear fields
- Make all fields private
- Optimize dynamic array growing
- Support custom fields

### Getting started

	# Grab the code from this repository and install the proto package.
	go get -u github.com/dropbox/goprotoc

### Running goprotoc

The compiler plugin, protoc-gen-dgo, will be installed in $GOBIN, defaulting to $GOPATH/bin. It must be in your $PATH for the protocol compiler, protoc, to find it.

Once the software is installed, there are two steps to using it. First you must compile the protocol buffer definitions and then import them, with the support library, into your program.

To compile the protocol buffer definition, run protoc with the --dgo_out parameter set to the directory you want to output the Go code to.

	protoc --dgo_out=. *.proto

If you are using any gogo.proto extensions you will need to specify the proto_path to include the descriptor.proto and gogo.proto. Located in github.com/dropbox/goprotoc/gogoproto and github.com/dropbox/goprotoc/protobuf respectively.

The proto package converts data structures to and from the
wire format of protocol buffers.  It works in concert with the
Go source code generated for .proto files by the protocol compiler.

### A summary of the properties of the goprotc protocol buffer interface:

TODO
