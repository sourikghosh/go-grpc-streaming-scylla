clean-files:
	rm -rf files/*

clean-pb:
	rm -rf internal/pb

gen:
	# protoc -I=$$PWD --go_out=$$PWD $$PWD/pkg/protos/*.proto
	protoc -I=$$PWD --go_out=$$PWD --go-grpc_out=$$PWD $$PWD/pkg/protos/*.proto

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

.PHONY: clean-pb gen server client