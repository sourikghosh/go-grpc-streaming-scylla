clean:
	rm -rf internal/pb

gen:
	# protoc -I=$$PWD --go_out=$$PWD $$PWD/pkg/protos/*.proto
	protoc -I=$$PWD --go_out=$$PWD --go-grpc_out=$$PWD $$PWD/pkg/protos/*.proto

server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

.PHONY: clean gen server client