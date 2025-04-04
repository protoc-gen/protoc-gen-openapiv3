package main

import (
	"github.com/protoc-gen/protoc-gen-openapiv3/openapiv3"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		openapiv3.GenerateFile(gen)
		gen.SupportedFeatures = gengo.SupportedFeatures
		return nil
	})
}
