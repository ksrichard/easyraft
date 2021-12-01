grpc:
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	go get -u google.golang.org/grpc
	protoc --go_out=plugins=grpc:. proto/*.proto

download:
	go mod tidy

install: download grpc