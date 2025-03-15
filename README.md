<div align="center">
    <h1>protoc-gen-openapiv3</h1>
</div>

<div align="center">

| [English](README.md) | [简体中文](docs/README_zh-CN.md) |

</div>

---

`protoc-gen-openapiv3` is a plugin for `protoc` designed to generate OpenAPI v3 descriptions from Protocol Buffer definitions. It simplifies the process of creating OpenAPI documentation by automatically converting your Protocol Buffer service definitions into OpenAPI specifications.

## Key Features
- **Automatic OpenAPI v3 Generation**: Seamlessly convert Protocol Buffer definitions to OpenAPI v3 specifications
- **Multi-language Support**: Generate documentation that can be used with various programming languages and frameworks
- **Customizable Output**: Configure the generated OpenAPI specifications through protobuf options
- **MIT Licensed**: Free to use, modify, and distribute under the MIT license

## Quick Start
Check out the [example](./example) directory for a quick start guide. Here's a brief overview:

### Definition
```protobuf
syntax = "proto3";

package trip.v1;

import "google/api/annotations.proto";
import "openapiv3/openapiv3.proto";

option go_package = "github.com/protoc-gen/protoc-gen-openapiv3/example/api/trip/v1;v1";

service TripService {
  rpc CreateTrip(CreateTripRequest) returns (CreateTripResponse) {
    option (google.api.http) = {
      post: "/api/v1/trips"
      body: "*"
    };
  }
}

message Trip {
  string id = 1 [(openapiv3.example) = {value: "680b81df-e966-4b51-a63f-1dfa749c04a5"}];
  string title = 2 [(openapiv3.example) = {value: "My Trip"}];
  string description = 3;
}

message CreateTripRequest {
  string title = 1;
  string description = 2;
}

message CreateTripResponse {
  Trip trip = 1;
}
```

### Installation
```bash
# Install the plugin
go install github.com/protoc-gen/protoc-gen-openapiv3

# Initialize development environment (optional)
make init
```

### Generation
You can generate OpenAPI v3 specifications in several ways:

1. Basic usage:
```bash
protoc --proto_path=. \
       --openapiv3_out=paths=source_relative:. \
       ./example/*.proto
```

2. With custom options:
```bash
protoc --proto_path=. \
       --proto_path=./third_party \
       --openapiv3_out=paths=source_relative:. \
       --openapiv3_opt=openapi_out_path=./example \
       --openapiv3_opt=servers='https://localhost:8000|Dev Server;https://localhost:9000|Prod Server' \
       ./example/*.proto
```

3. Using Makefile:
```bash
# Generate all (includes example and core files)
make all

# Generate only example files
make example
```

### Usage
The generated OpenAPI v3 specification can be used with any OpenAPI-compatible tool or framework. We provide two example HTML viewers in the example directory:

1. Swagger UI: Open `example/swagger.html` in your browser
2. RapiDoc: Open `example/index.html` in your browser

Both viewers provide an interactive interface to explore and test your API endpoints.

## License
This project is licensed under the MIT License. See [LICENSE](./LICENSE) for the full license text.
