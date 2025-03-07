.PHONY: init
# init env
init:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3

.PHONY: openapiv3
# generate openapiv3
openapiv3:
	protoc --proto_path=./openapiv3 \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:./openapiv3 \
		   ./openapiv3/*.proto

.PHONY: example
# generate example
example:
	go install . && \
	protoc --proto_path=. \
		   --proto_path=./third_party \
		   --openapiv3_out=paths=source_relative:. \
		   --openapiv3_opt=openapi_out_path=./example \
		   --openapiv3_opt=servers='https://localhost:8000|Dev Server;https://localhost:9000|Prod Server' \
		   ./example/*.proto

.PHONY: all
# generate all
all:
	make openapiv3;
	make example;
	go mod tidy;
