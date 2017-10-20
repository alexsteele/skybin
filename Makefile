
proto: core/proto
	cd core/proto && protoc *.proto --go_out=plugins=grpc:.

