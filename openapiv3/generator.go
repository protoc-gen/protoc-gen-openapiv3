package openapiv3

import (
	"fmt"
	"os"
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"gopkg.in/yaml.v3"
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
	openAPI["components"] = components

	// Traverse all proto files
	for _, file := range gen.Files {
		if file.Generate {
			// Traverse each service in the file
			for _, service := range file.Services {
				// Create a base path for the service
				servicePath := "/" + service.GoName
				for _, method := range service.Methods {
					// Generate OpenAPI path for each method under the service
					methodPath := fmt.Sprintf("%s/%s", servicePath, method.GoName)
					operation := map[string]interface{}{
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
						"post": operation,
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
