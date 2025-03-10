package openapiv3

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
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
