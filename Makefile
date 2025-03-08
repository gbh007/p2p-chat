.PHONY: proto
proto:
	protoc -I=. --go_out=. --go-grpc_out=. proto/server.proto

.PHONY: install-proto
install-proto:
	sudo apt install protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@lates