package openapiv3

import (
	"fmt"
	"github.com/protoc-gen/protoc-gen-openapiv3/pkg/helper"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

// GenerateFile traverses all proto files and generates the OpenAPI specification file
func GenerateFile(gen *protogen.Plugin) {
	// Basic structure of the OpenAPI specification
	openAPI := make(map[string]any)
	openAPI["openapi"] = "3.0.0"
	openAPI["info"] = map[string]any{
		"title":       "Generated API",
		"description": "API generated from protobufs",
		"version":     "1.0.0",
	}

	servers := parseServersOption(gen)
	if len(servers) > 0 {
		openAPI["servers"] = servers
	}

	// Middle part for paths
	paths := make(map[string]map[string]any)
	openAPI["paths"] = paths

	// Components part, will be added at the end
	components := make(map[string]any)
	commonResp := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"code": map[string]any{
				"type": "integer",
			},
			"message": map[string]any{
				"type": "string",
			},
			"details": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"@type": map[string]any{
							"type": "string",
						},
						"reason": map[string]any{
							"type": "string",
						},
						"domain": map[string]any{
							"type": "string",
						},
						"metadata": map[string]any{
							"type": "object",
						},
					},
				},
			},
		},
		"example": map[string]any{
			"code":    3,
			"message": "must be at least 11 characters long",
			"details": []map[string]any{
				{
					"@type":  "type.googleapis.com/google.rpc.ErrorInfo",
					"reason": "INVALID_PARAMETERS",
					"domain": "",
					"metadata": map[string]any{
						"field": "phoneNumber",
					},
				},
			},
		},
	}
	// {"code":3, "message":"Request is invalid", "details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo", "reason":"BAD_REQUEST", "domain":"", "metadata":{}}]}%
	components["schemas"] = map[string]any{
		"BadRequest":          commonResp,
		"Unauthorized":        commonResp,
		"InternalServerError": commonResp,
	}
	components["securitySchemes"] = map[string]any{
		"BearerAuth": map[string]any{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		},
	}
	openAPI["components"] = components

	// Security part
	openAPI["security"] = []map[string]any{
		{
			"BearerAuth": []any{},
		},
	}

	allTags := map[string]string{}

	// Traverse all proto files
	for _, file := range gen.Files {
		if file.Generate {
			// Traverse each service in the file
			for _, service := range file.Services {
				svcName := GetServiceName(service)
				allTags[svcName] = GetServiceDescription(service)
				for _, method := range service.Methods {
					// Generate OpenAPI path for each method under the service
					operation := map[string]any{
						"tags":        []string{svcName},
						"operationId": fmt.Sprintf("%s_%s", service.GoName, method.GoName),
						"requestBody": getRequestBody(method.Input),
						"responses":   getResponseBody(method.Output),
					}

					// Check if skip_token is true
					methodOpts := proto.GetExtension(method.Desc.Options(), E_Method).(*Method)
					if methodOpts != nil && methodOpts.SkipToken {
						operation["security"] = []map[string]any{}
					}

					// Add operation to paths
					methodPath, httpMethod := helper.GetHttpMethodAndPath(method)
					if _, ok := paths[methodPath]; !ok {
						paths[methodPath] = make(map[string]any)
					}
					paths[methodPath][httpMethod] = operation

					// Generate schema for Input and Output
					addMessageSchema(openAPI, method.Input)
					addMessageSchema(openAPI, method.Output)
				}
			}
		}
	}

	// Tags specification:
	// https://swagger.io/docs/specification/v3_0/grouping-operations-with-tags/
	tags := make([]map[string]any, 0, len(allTags))
	for tag, desc := range allTags {
		tags = append(tags, map[string]any{
			"name":        tag,
			"description": desc,
		})
	}
	openAPI["tags"] = tags

	// Generate OpenAPI YAML file
	openAPIDocument, err := yaml.Marshal(openAPI)
	if err != nil {
		fmt.Println("Error marshalling OpenAPI document:", err)
		return
	}

	// Save to file
	err = os.WriteFile(getOutputFilename(gen), openAPIDocument, 0644)
	if err != nil {
		fmt.Println("Error writing OpenAPI file:", err)
	}
}

// addMessageSchema adds proto message types to OpenAPI components
func addMessageSchema(openAPI map[string]any, message *protogen.Message) {
	schemaName := helper.GetSchemaName(message)
	if components, ok := openAPI["components"].(map[string]any); ok {
		if schemas, ok := components["schemas"].(map[string]any); ok {
			// Construct schema
			schema := make(map[string]any)
			schema["type"] = "object"
			properties := make(map[string]any)
			example := make(map[string]any)

			// Traverse fields and generate properties
			for _, field := range message.Fields {
				property := make(map[string]any)
				switch field.Desc.Kind() {
				case protoreflect.BoolKind:
					property["type"] = "boolean"
					example[field.Desc.JSONName()] = getExample(field, true)
				case protoreflect.EnumKind:
					property["type"] = "string"
					property["format"] = "enum"
					// Enum specification:
					// https://swagger.io/docs/specification/v3_0/data-models/enums/
					property["enum"] = helper.GetEnumValues(field.Enum)
					example[field.Desc.JSONName()] = getExample(field, property["enum"].([]string)[0])
				case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
					protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
					property["type"] = "integer"
					property["format"] = "int32"
					example[field.Desc.JSONName()] = getExample(field, 0)
				case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
					protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
					property["type"] = "integer"
					property["format"] = "int64"
					example[field.Desc.JSONName()] = getExample(field, 0)
				case protoreflect.FloatKind:
					property["type"] = "number"
					property["format"] = "float"
					example[field.Desc.JSONName()] = getExample(field, 0.0)
				case protoreflect.DoubleKind:
					property["type"] = "number"
					property["format"] = "double"
					example[field.Desc.JSONName()] = getExample(field, 0.0)
				case protoreflect.StringKind:
					property["type"] = "string"
					example[field.Desc.JSONName()] = getExample(field, "")
				case protoreflect.BytesKind:
					property["type"] = "string"
					property["format"] = "byte" // Or use "binary" if needed for base64 encoding
					example[field.Desc.JSONName()] = getExample(field, "")
				case protoreflect.MessageKind, protoreflect.GroupKind:
					if helper.GetSchemaName(field.Message) == "google.protobuf.Timestamp" {
						// This is google.protobuf.Timestamp, treat it as a date-time string
						property["type"] = "integer"
						property["format"] = "int32"
						example[field.Desc.JSONName()] = getExample(field, 1741589979)
					} else {
						// Otherwise, treat it as a regular message and add a reference to the schema
						addMessageSchema(openAPI, field.Message)
						property["$ref"] = fmt.Sprintf("#/components/schemas/%s", helper.GetSchemaName(field.Message))
					}
				default:
					property["type"] = "string"
					example[field.Desc.JSONName()] = ""
				}

				// Add property to schema
				properties[field.Desc.JSONName()] = property
			}

			// Add generated properties to schema
			schema["properties"] = properties
			if len(example) > 0 {
				schema["example"] = example
			}

			// Add schema to components/schemas
			schemas[schemaName] = schema
		}
	}
}

// getOutputFilename extracts the output file path from plugin options
func getOutputFilename(gen *protogen.Plugin) string {
	parts := strings.Split(gen.Request.GetParameter(), ",")

	filename := "openapi.yaml"
	// TODO: is it possible to read the --openapiv3_out=paths=source_relative:./example from the plugin options?
	for _, part := range parts {
		if strings.HasPrefix(part, "openapi_out_path=") {
			return path.Join(strings.TrimPrefix(part, "openapi_out_path="), filename)
		}
	}

	return filename
}

// parseServersOption parses the servers option from the plugin options
func parseServersOption(gen *protogen.Plugin) []map[string]any {
	parts := strings.Split(gen.Request.GetParameter(), ",")
	servers := make([]map[string]any, 0)
	for _, part := range parts {
		if strings.HasPrefix(part, "servers=") {
			for _, server := range strings.Split(strings.TrimPrefix(part, "servers="), ";") {
				info := strings.Split(server, "|")
				if len(info) == 2 {
					servers = append(servers, map[string]any{
						"url":         info[0],
						"description": info[1],
					})
				} else {
					servers = append(servers, map[string]any{
						"url": info[0],
					})
				}
			}
		}
	}
	return servers
}

func getRequestBody(message *protogen.Message) map[string]any {
	return map[string]any{
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{
					"$ref": fmt.Sprintf("#/components/schemas/%s", helper.GetSchemaName(message)),
				},
			},
		},
	}
}

func getResponseBody(message *protogen.Message) map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "OK",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": fmt.Sprintf("#/components/schemas/%s", helper.GetSchemaName(message)),
					},
				},
			},
		},
		"400": map[string]any{
			"description": "Bad Request",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/BadRequest",
					},
				},
			},
		},
		"401": map[string]any{
			"description": "Unauthorized",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/Unauthorized",
					},
				},
			},
		},
		"500": map[string]any{
			"description": "Internal Server Error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/InternalServerError",
					},
				},
			},
		},
	}
}
