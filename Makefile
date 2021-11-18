clean-files:
	rm -rf files/*

clean-pb:
	rm -rf internal/pb

gen:
	protoc -I=$$PWD --go_out=$$PWD --go-grpc_out=$$PWD $$PWD/pkg/protos/*.proto

server:
	@echo "Running scyllaDB in cluster"
	docker-compose up -d
	@echo "waiting cluster to get ready for connection"
	sleep 130
	@echo "Running DB migration"
	docker exec go-grpc-streaming-scylla_scylla-node1_1 cqlsh -f /init.txt 
	sleep 20
	@echo "Running Apex Server"
	go run cmd/server/main.go -port 1500

client:
	go install cmd/client/apex.go

.PHONY: clean-pb gen server client