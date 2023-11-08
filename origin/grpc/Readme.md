go get -u google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go

go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

protoc --proto_path=./proto ./proto/*.proto --plugin=$(go env GOPATH)/bin/protoc-gen-go --go_out=./

protoc --proto_path=./proto ./proto/*.proto --plugin=$(go env GOPATH)/bin/protoc-gen-go-grpc --go-grpc_out=./
