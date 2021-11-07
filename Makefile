clean:
	rm -rf internal/pb

gen:
	protoc -I=$$PWD --go_out=$$PWD $$PWD/pkg/protos/*.proto