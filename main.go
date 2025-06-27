package main

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protoc-gen/protoc-gen-openapiv3/openapiv3"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			openapiv3.GenerateFile(gen, f)
		}
		return nil
	})
}
