package main

import (
	"github.com/protoc-gen/protoc-gen-openapiv3/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		openapiv3.GenerateFile(gen)
		return nil
	})
}
