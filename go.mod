module github.com/echovisionlab/geul-event-contracts

go 1.26.6

require (
	connectrpc.com/connect v1.20.0
	google.golang.org/protobuf v1.36.12
)

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)
