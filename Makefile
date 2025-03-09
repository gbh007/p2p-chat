.PHONY: proto
proto:
	protoc -I=. --go_out=. --go-grpc_out=. proto/server.proto

.PHONY: install-proto
install-proto:
	sudo apt install protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: client
client:
	go run cmd/client/main.go

.PHONY: server
server:
	go run cmd/server/main.go