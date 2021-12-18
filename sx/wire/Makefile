all:
	protoc -I=. -I=${GOPATH}/src -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	-I=${GOPATH}/src/github.com/gogo/protobuf/gogoproto --gogofast_out=. wire.proto