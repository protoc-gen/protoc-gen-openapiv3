package openapiv3

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
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

func isUUIDByName(name string) bool {
	return name == "id" || strings.HasSuffix(name, "Id")
}
