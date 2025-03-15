<div align="center">
    <h1>protoc-gen-openapiv3</h1>
</div>

<div align="center">

| [English](../README.md) | [简体中文](README_zh-CN.md) |

</div>

---

`protoc-gen-openapiv3` 是一个用于 `protoc` 的插件，旨在从 Protocol Buffer 定义生成 OpenAPI v3 描述文档。它通过自动将您的 Protocol Buffer 服务定义转换为 OpenAPI 规范，简化了创建 OpenAPI 文档的过程。

## 关键特性
- **自动生成 OpenAPI v3**：无缝将 Protocol Buffer 定义转换为 OpenAPI v3 规范
- **多语言支持**：生成的文档可用于各种编程语言和框架
- **可自定义输出**：通过 protobuf 选项配置生成的 OpenAPI 规范
- **MIT 许可**：基于 MIT 许可证，可以自由使用、修改和分发

## 快速开始
查看 [example](../example) 目录以获取快速入门指南。以下是简要概述：

### 定义
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

### 安装
```bash
# 安装插件
go install github.com/protoc-gen/protoc-gen-openapiv3

# 初始化开发环境（可选）
make init
```

### 生成
您可以通过以下几种方式生成 OpenAPI v3 规范：

1. 基本用法：
```bash
protoc --proto_path=. \
       --openapiv3_out=paths=source_relative:. \
       ./example/*.proto
```

2. 使用自定义选项：
```bash
protoc --proto_path=. \
       --proto_path=./third_party \
       --openapiv3_out=paths=source_relative:. \
       --openapiv3_opt=openapi_out_path=./example \
       --openapiv3_opt=servers='https://localhost:8000|Dev Server;https://localhost:9000|Prod Server' \
       ./example/*.proto
```

3. 使用 Makefile：
```bash
# 生成所有文件（包括示例和核心文件）
make all

# 仅生成示例文件
make example
```

### 使用
生成的 OpenAPI v3 规范可以与任何兼容 OpenAPI 的工具或框架一起使用。我们在示例目录中提供了两个 HTML 查看器：

1. Swagger UI：在浏览器中打开 `example/swagger.html`
2. RapiDoc：在浏览器中打开 `example/index.html`

这两个查看器都提供了交互式界面，可以用于探索和测试您的 API 端点。

## 许可
该项目采用 MIT 许可证。查看 [LICENSE](../LICENSE) 获取完整的许可证文本。 