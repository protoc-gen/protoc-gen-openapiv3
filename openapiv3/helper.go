package openapiv3

import (
	"fmt"
	"github.com/protoc-gen/protoc-gen-openapiv3/pkg/helper"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strconv"
)

func GetServiceName(svc *protogen.Service) string {
	svcOpts := proto.GetExtension(svc.Desc.Options(), E_Service).(*Service)
	if svcOpts != nil && svcOpts.GetName() != "" {
		return svcOpts.GetName()
	}

	return svc.GoName
}

func GetServiceDescription(svc *protogen.Service) string {
	svcOpts := proto.GetExtension(svc.Desc.Options(), E_Service).(*Service)
	if svcOpts != nil {
		return svcOpts.GetDescription()
	}

	return ""
}

type openAPITypes interface {
	~int | ~string | ~bool | ~float64 | ~float32
}

func getExample[T openAPITypes](field *protogen.Field, defValue T) T {
	opt := proto.GetExtension(field.Desc.Options(), E_Example).(*Example)
	if opt == nil {
		return defValue
	}

	val := opt.GetValue()

	switch any(defValue).(type) {
	case int:
		if i, err := strconv.Atoi(val); err == nil {
			return any(i).(T)
		}
		return defValue
	case string:
		return any(val).(T)
	case bool:
		if b, err := strconv.ParseBool(val); err == nil {
			return any(b).(T)
		}
		return defValue
	case float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return any(f).(T)
		}
		return defValue
	case float32:
		if f, err := strconv.ParseFloat(val, 32); err == nil {
			return any(float32(f)).(T)
		}
		return defValue
	default:
		return defValue
	}
}

type nestedMessageCallback func(*protogen.Message)

func GetPropertyAndExample(field *protogen.Field, nestedMessageCallback nestedMessageCallback) (map[string]interface{}, any) {
	var (
		property = make(map[string]interface{})
		example  any
	)

	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		property["type"] = "boolean"
		example = getExample(field, true)
	case protoreflect.EnumKind:
		property["type"] = "string"
		property["format"] = "enum"
		// Enum specification:
		// https://swagger.io/docs/specification/v3_0/data-models/enums/
		property["enum"] = helper.GetEnumValues(field.Enum)
		example = getExample(field, property["enum"].([]string)[0])
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		property["type"] = "integer"
		property["format"] = "int32"
		example = getExample(field, 0)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		property["type"] = "integer"
		property["format"] = "int64"
		example = getExample(field, 0)
	case protoreflect.FloatKind:
		property["type"] = "number"
		property["format"] = "float"
		example = getExample(field, 0.0)
	case protoreflect.DoubleKind:
		property["type"] = "number"
		property["format"] = "double"
		example = getExample(field, 0.0)
	case protoreflect.StringKind:
		property["type"] = "string"
		example = getExample(field, "")
	case protoreflect.BytesKind:
		property["type"] = "string"
		property["format"] = "byte" // Or use "binary" if needed for base64 encoding
		example = getExample(field, "")
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if helper.GetSchemaName(field.Message) == "google.protobuf.Timestamp" {
			// This is google.protobuf.Timestamp, treat it as a date-time string
			property["type"] = "integer"
			property["format"] = "int32"
			example = getExample(field, 1741589979)
		} else {
			// Otherwise, treat it as a regular message and add a reference to the schema
			if nestedMessageCallback != nil {
				nestedMessageCallback(field.Message)
				property["$ref"] = fmt.Sprintf("#/components/schemas/%s", helper.GetSchemaName(field.Message))
			}
		}
	default:
		property["type"] = "string"
		example = ""
	}

	if field.Desc.Cardinality() == protoreflect.Repeated {
		property = map[string]any{
			"type":  "array",
			"items": property,
		}
	}

	return property, example
}
