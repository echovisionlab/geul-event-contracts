module github.com/echovisionlab/geul-event-contracts

go 1.27.3

require (
	connectrpc.com/connect v1.21.1
	google.golang.org/protobuf v1.36.12
)

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)
