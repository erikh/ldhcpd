package proto

//go:generate protoc --proto_path=.:$GOPATH/src --go_out=plugins=grpc:. control.proto
