.PHONY: compileAll
compileAll: v1.name v1.order v1.brokerjwt

.PHONY: v1.name
v1.name:
	protoc ./api/v1/name/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

.PHONY: v1.order
v1.order:
	protoc ./api/v1/order/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

.PHONY: v1.brokerjwt
v1.brokerjwt:
	protoc ./api/v1/brokerjwt/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

