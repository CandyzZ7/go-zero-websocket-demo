.PHONY: proto
proto:
	protoc --go_out=./internal  --go-grpc_out=. ./internal/proto/*.proto
