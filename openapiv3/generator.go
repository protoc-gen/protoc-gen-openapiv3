package openapiv3

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/protoc-gen/protoc-gen-openapiv3/pkg/helper"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// GenerateFile traverses all proto files and generates the OpenAPI specification file
func GenerateFile(gen *protogen.Plugin) {
	paths := make(map[string]map[string]any)
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

	// Basic structure of the OpenAPI specification
	openAPI := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":       "Generated API",
			"description": "API generated from protobufs",
			"version":     "1.0.0",
		},
		"security": []map[string]any{
			{
				"BearerAuth": []any{},
			},
		},
		"paths": paths,
		"components": map[string]any{
			"schemas": map[string]any{
				"BadRequest":          commonResp,
				"Unauthorized":        commonResp,
				"InternalServerError": commonResp,
			},
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
	}

	servers := parseServersOption(gen)
	if len(servers) > 0 {
		openAPI["servers"] = servers
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
						"responses":   getResponseBody(method.Output),
					}

					// Check if skip_token is true
					methodOpts := proto.GetExtension(method.Desc.Options(), E_Method).(*Method)
					if methodOpts != nil && methodOpts.SkipToken {
						operation["security"] = []map[string]any{}
					}

					addMessageSchema(openAPI, method.Output)
					methodPath, httpMethod, bindings := helper.GetHttpMethodAndPath(method)
					if httpMethod == "post" || httpMethod == "put" || httpMethod == "patch" {
						operation["requestBody"] = getRequestBody(method.Input)
						addMessageSchema(openAPI, method.Input)
					}

					parameters := extractPathParameters(method.Input, methodPath, bindings)
					if len(parameters) > 0 {
						operation["parameters"] = parameters
					}

					if _, ok := paths[methodPath]; !ok {
						paths[methodPath] = make(map[string]any)
					}
					paths[methodPath][httpMethod] = operation
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
	// sort tags by name
	sort.Slice(tags, func(i, j int) bool {
		return tags[i]["name"].(string) < tags[j]["name"].(string)
	})
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
			examples := make(map[string]any)

			// Traverse fields and generate properties
			for _, field := range message.Fields {
				property, example := GetPropertyAndExample(field, func(message *protogen.Message) {
					addMessageSchema(openAPI, message)
				})

				examples[field.Desc.JSONName()] = example
				properties[field.Desc.JSONName()] = property
			}

			// Add generated properties to schema
			schema["properties"] = properties
			if len(examples) > 0 {
				schema["example"] = examples
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
		"required": true,
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

type extractTarget interface {
	[]string | string
}

func extractKeys[T extractTarget](s T) map[string]struct{} {
	keys := make(map[string]struct{})
	switch v := any(s).(type) {
	case []string:
		for _, vv := range v {
			// /api/v1/trips?page={page}&size={size}
			parts := strings.Split(vv, "{")
			for _, part := range parts[1:] {
				end := strings.Index(part, "}")
				if end != -1 {
					keys[part[:end]] = struct{}{}
				}
			}
		}
	case string:
		parts := strings.Split(v, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				keys[part[1:len(part)-1]] = struct{}{}
			}
		}
	}
	return keys
}

func extractPathParameters(message *protogen.Message, uri string, bindings []string) []map[string]any {
	var parameters []map[string]any
	pathKeys := extractKeys(uri)
	for _, field := range message.Fields {
		// only add path parameters that are in the bindings
		if _, ok := pathKeys[string(field.Desc.Name())]; !ok {
			continue
		}
		params := map[string]any{
			"name":     field.Desc.JSONName(),
			"in":       "path",
			"required": true,
		}
		property, example := GetPropertyAndExample(field, nil)
		params["schema"] = property
		params["example"] = example
		parameters = append(parameters, params)
	}

	queryKeys := extractKeys(bindings)
	for _, field := range message.Fields {
		// only add query parameters that are in the bindings
		if _, ok := queryKeys[string(field.Desc.Name())]; !ok {
			continue
		}
		params := map[string]any{
			"name":     field.Desc.JSONName(),
			"in":       "query",
			"required": false,
		}
		property, example := GetPropertyAndExample(field, nil)
		params["schema"] = property
		params["example"] = example
		parameters = append(parameters, params)
	}
	return parameters
}
