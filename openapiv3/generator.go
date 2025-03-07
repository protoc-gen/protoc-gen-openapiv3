package openapiv3

import (
	"fmt"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

// GenerateFile traverses all proto files and generates the OpenAPI specification file
func GenerateFile(gen *protogen.Plugin) {
	// Basic structure of the OpenAPI specification
	openAPI := make(map[string]interface{})
	openAPI["openapi"] = "3.0.0"
	openAPI["info"] = map[string]interface{}{
		"title":       "Generated API",
		"description": "API generated from protobufs",
		"version":     "1.0.0",
	}

	// Middle part for paths
	paths := make(map[string]interface{})
	openAPI["paths"] = paths

	// Components part, will be added at the end
	components := make(map[string]interface{})
	components["schemas"] = make(map[string]interface{})
	components["securitySchemes"] = map[string]interface{}{
		"BearerAuth": map[string]interface{}{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		},
	}
	openAPI["components"] = components

	// Security part
	openAPI["security"] = []map[string]interface{}{
		{
			"BearerAuth": []interface{}{},
		},
	}

	// Traverse all proto files
	for _, file := range gen.Files {
		if file.Generate {
			// Traverse each service in the file
			for _, service := range file.Services {
				for _, method := range service.Methods {
					httpRule := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
					var methodPath string
					var httpMethod string
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
					// Generate OpenAPI path for each method under the service
					operation := map[string]interface{}{
						"tags":        []string{service.GoName},
						"operationId": fmt.Sprintf("%s_%s", service.GoName, method.GoName),
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "Successful Response",
								"content": map[string]interface{}{
									"application/json": map[string]interface{}{
										"schema": map[string]interface{}{
											"$ref": fmt.Sprintf("#/components/schemas/%s", method.Input.GoIdent.GoName),
										},
									},
								},
							},
						},
					}

					// Check if openapiv3.skip_token is true
					skipToken := proto.GetExtension(method.Desc.Options(), E_SkipToken).(bool)
					if skipToken {
						operation["security"] = []map[string]interface{}{}
					}

					// Generate OpenAPI request body for each method
					operation["requestBody"] = map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": fmt.Sprintf("#/components/schemas/%s", method.Input.GoIdent.GoName),
								},
							},
						},
					}

					// Add operation to paths
					paths[methodPath] = map[string]interface{}{
						httpMethod: operation,
					}

					// Generate schema for Input and Output
					addMessageSchema(openAPI, method.Input)
					addMessageSchema(openAPI, method.Output)
				}
			}
		}
	}

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
func addMessageSchema(openAPI map[string]interface{}, message *protogen.Message) {
	// Get message name
	schemaName := message.GoIdent.GoName
	if components, ok := openAPI["components"].(map[string]interface{}); ok {
		if schemas, ok := components["schemas"].(map[string]interface{}); ok {
			// Construct schema
			schema := make(map[string]interface{})
			schema["type"] = "object"
			properties := make(map[string]interface{})

			// Traverse fields and generate properties
			for _, field := range message.Fields {
				property := make(map[string]interface{})
				property["type"] = "string" // Assume string, can be adjusted based on actual type

				// Add property to schema
				properties[field.GoName] = property
			}

			// Add generated properties to schema
			schema["properties"] = properties

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
