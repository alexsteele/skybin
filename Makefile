
proto: core/pb
	cd core/pb && protoc *.proto --go_out=plugins=grpc:.

