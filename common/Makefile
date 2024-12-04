customer_gen:
	@protoc \
		--proto_path=protobuf "protobuf/customer.proto" \
		./protobuf/customer.proto \
		--go_out=.././services/customer/grpc-handler/pb --go_opt=paths=source_relative \
		--go-grpc_out=.././services/customer/grpc-handler/pb --go-grpc_opt=paths=source_relative

product_gen:
	@protoc \
		--proto_path=protobuf "protobuf/product.proto" \
		./protobuf/product.proto \
		--go_out=.././services/inventory/grpc/pb --go_opt=paths=source_relative \
		--go-grpc_out=.././services/inventory/grpc/pb --go-grpc_opt=paths=source_relative


