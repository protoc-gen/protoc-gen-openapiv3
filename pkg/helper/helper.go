package helper

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
)

func GetSchemaName(message *protogen.Message) string {
	packageName := string(message.Desc.ParentFile().Package())
	return fmt.Sprintf("%s.%s", packageName, message.GoIdent.GoName)
}
