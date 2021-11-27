clean-files:
	rm -rf files/*

clean-pb:
	rm -rf internal/pb

gen:
	protoc -I=$$PWD --go_out=$$PWD --go-grpc_out=$$PWD $$PWD/pkg/protos/*.proto

server:
	@echo "Running Apex Server..."
	mkdir -p files
	go run cmd/server/main.go -port 1500

client:
	@echo "Installing Apex cli..."
	go install cmd/client/apex.go

.PHONY: clean-pb gen server client