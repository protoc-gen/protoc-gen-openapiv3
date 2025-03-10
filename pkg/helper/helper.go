package helper

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
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
