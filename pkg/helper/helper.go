package helper

import (
	"fmt"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func GetSchemaName(message *protogen.Message) string {
	packageName := string(message.Desc.ParentFile().Package())
	return fmt.Sprintf("%s.%s", packageName, message.GoIdent.GoName)
}

func GetEnumValues(enum *protogen.Enum) []string {
	values := make([]string, 0, len(enum.Values))
	for _, v := range enum.Values {
		values = append(values, string(v.Desc.Name()))
	}
	return values
}

func GetHttpMethodAndPath(method *protogen.Method) (methodPath string, httpMethod string) {
	httpRule := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
	if httpRule != nil {
		switch pattern := httpRule.Pattern.(type) {
		case *annotations.HttpRule_Post:
			methodPath = pattern.Post
			httpMethod = "post"
		case *annotations.HttpRule_Get:
			methodPath = pattern.Get
			httpMethod = "get"
		case *annotations.HttpRule_Put:
			methodPath = pattern.Put
			httpMethod = "put"
		case *annotations.HttpRule_Delete:
			methodPath = pattern.Delete
			httpMethod = "delete"
		case *annotations.HttpRule_Patch:
			methodPath = pattern.Patch
			httpMethod = "patch"
		}
	}
	return
}
