.PHONY: proto
# Generate Go proto
proto:
	protoc --go_out=./internal  --go-grpc_out=. ./internal/proto/*.proto
.PHONY: gen
# Generate Go generated
gen:
	go generate ./...
