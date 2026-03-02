# http-little-toy

## 简介 Introduction

这是一个简单而强大的HTTP并发测试工具，使用Go语言编写。如果您觉得好用，请点个Star⭐️！
This is a simple yet powerful HTTP concurrency testing tool written in Go. Star ⭐️ this repository if you find it useful!
 
## 主要特性 Key Features

- **并发测试 Concurrent Testing**: 支持多线程并发HTTP请求 / Support multi-threaded concurrent HTTP requests
- **协议支持 Protocol Support**: HTTP/1.1 和 HTTP/2 协议 / HTTP/1.1 and HTTP/2 protocols
- **灵活配置 Flexible Configuration**: 丰富的参数配置选项 / Rich parameter configuration options
- **详细统计 Detailed Statistics**: 全面的性能指标和统计数据 / Comprehensive performance metrics and statistics
- **安全支持 Security Support**: TLS/SSL证书认证支持 / TLS/SSL certificate authentication support
- **跨平台 Cross-platform**: 支持Windows、Linux、macOS / Support Windows, Linux, macOS

## 安装指南 Installation Guide

### 方法一：Go Install（推荐） Method 1: Go Install (Recommended)
```bash
go install github.com/leihenshang/http-little-toy@latest
```
确保您的 `GOPATH/bin` 在系统PATH中。
Make sure your `GOPATH/bin` is in your system PATH.

### 方法二：手动编译 Method 2: Manual Compilation
```bash
git clone https://github.com/leihenshang/http-little-toy.git
cd http-little-toy
go build -o http-little-toy
./http-little-toy -h
```

## 快速开始 Quick Start

### 基础使用 Basic Usage
```bash
# 简单GET请求测试 Simple GET request test
http-little-toy -u http://example.com -t 10 -d 30

# 带自定义头部的POST请求 POST request with custom headers
http-little-toy -u http://api.example.com/users -m POST \
  -header "Content-Type: application/json" \
  -header "Authorization: Bearer token123" \
  -body '{"name":"test"}' \
  -t 20 -d 60
```

### 输出控制 Output Control
```bash
# 输出JSON格式结果 Output results in JSON format
http-little-toy -u http://example.com -t 10 -d 30 -format json

# 将结果保存到文件 Save results to file
http-little-toy -u http://example.com -t 10 -d 30 -resFile result.txt

# JSON格式并保存到文件 JSON format and save to file
http-little-toy -u http://example.com -t 10 -d 30 -format json -resFile result.json
```

## 命令行参数 Command Line Options

### 基本参数 Basic Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-u` | 测试目标URL（必填）Target URL (required) | - | `-u http://example.com` |
| `-m` | HTTP请求方法 HTTP request method | GET | `-m POST` |
| `-t` | 并发线程数 Number of concurrent threads | 10 | `-t 50` |
| `-d` | 测试持续时间(秒) Test duration (seconds) | 10 | `-d 60` |
| `-timeout` | 单个请求超时时间(秒) Request timeout (seconds) | 10 | `-timeout 30` |
| `-header` | 自定义HTTP头部（可多次使用）Custom HTTP headers (can be used multiple times) | [] | `-header "Key: Value"` |
| `-body` | 请求体内容 Request body content | "" | `-body '{"data":"test"}'` |

### 连接与协议参数 Connection & Protocol Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-keepAlive` | 启用HTTP长连接（提高性能）Enable HTTP keep-alive (improves performance) | true | `-keepAlive=false` |
| `-compression` | 启用HTTP压缩传输 Enable HTTP compression | true | `-compression=false` |
| `-h2` | 使用HTTP/2协议 Use HTTP/2 protocol | false | `-h2=true` |
| `-allowRedirects` | 允许HTTP重定向 Allow HTTP redirects | true | `-allowRedirects=false` |

### TLS/SSL安全参数 TLS/SSL Security Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-skipVerify` | 跳过TLS证书验证（不安全）Skip TLS certificate verification (insecure) | false | `-skipVerify=true` |
| `-clientCert` | 客户端证书文件路径 Client certificate file path | "" | `-clientCert cert.pem` |
| `-clientKey` | 客户端私钥文件路径 Client private key file path | "" | `-clientKey key.pem` |
| `-caCert` | CA证书文件路径 CA certificate file path | "" | `-caCert ca.pem` |

### 输出参数 Output Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-resFile` | 结果输出文件路径 Result output file path | "" | `-resFile result.txt` |
| `-format` | 输出格式（raw/json/csv）Output format (raw/json/csv) | raw | `-format json` |

### 帮助参数 Help Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-h` | 显示帮助信息 Show help information | false | `-h` |
| `-v` | 显示版本信息 Show version information | false | `-v` |

## 跨平台编译 Cross-platform Compilation

### Linux & macOS 编译 Build
```bash
# Linux编译 Linux build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/http-little-toy

# Windows编译 Windows build
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/http-little-toy.exe

# macOS编译 macOS build
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/http-little-toy
```

### Windows CMD 编译 Build
```cmd
REM macOS编译 macOS build
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o bin/http-little-toy

REM Linux编译 Linux build
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/http-little-toy

REM Windows编译 Windows build
go build -o bin/http-little-toy.exe
```

## 使用示例 Usage Examples

### 基础测试 Basic Tests
```bash
# 简单并发测试 Simple concurrency test
./http-little-toy -d 10 -t 80 -u http://127.0.0.1:9090

# 带自定义头部测试 Test with custom headers
go run . -u http://localhost:9090 -header aaa:bbbb -header ccc:ddd -body "test data" -d 10 -t 10
```

### 高级测试 Advanced Tests
```bash
# 复杂API测试（JSON负载） Complex API test with JSON payload
./http-little-toy -u http://localhost:2025/api/users/list -m POST \
  -header "token: itoken_1754640941655527771" \
  -header "Content-Type: application/json" \
  -body '{
    "pageSize": 100,
    "page": 1
  }' -d 60 -t 30
```

### HTTPS证书认证测试 HTTPS with Certificate Authentication
```bash
# HTTPS测试（客户端证书） HTTPS test with client certificates
./http-little-toy -u https://secure-api.com \
  -clientCert client.crt \
  -clientKey client.key \
  -caCert ca.crt \
  -skipVerify=false \
  -d 30 -t 20
```

## 输出结果解读 Output Interpretation

工具提供全面的测试结果，包括：
The tool provides comprehensive test results including:

- **成功/失败次数 Success/Failure Count**: 成功和失败的请求数量 / Number of successful and failed requests
- **吞吐量 Throughput**: 每秒请求数(RPS) / Requests per second (RPS)
- **传输速率 Transfer Rate**: 数据传输速度(KB/s) / Data transfer speed (KB/s)
- **响应时间 Response Time**: 平均、最小和最大请求时间 / Average, minimum, and maximum request times
- **总数据量 Total Data**: 测试期间传输的数据总量 / Amount of data transferred during the test

示例输出 Example output:
```
use 10 coroutines,duration 30 seconds.
---------------stats---------------
Test Results:
  Success requests: 2847
  Failed requests: 3
  Total data received: 142.50 KB
  Requests per second: 94.90 RPS
  Transfer rate: 4.75 KB/s
  Average request time: 105ms
  Slowest request: 342ms
  Fastest request: 45ms
  Actual test duration: 30.001s
```

## 技术要求 Technical Requirements

- **Go版本 Go Version**: 1.22或更高版本 / 1.22 or higher
- **依赖库 Dependencies**: 
  - `golang.org/x/net/http2`
  - 其余均为标准库 / Standard library only otherwise

## 性能优化建议 Performance Optimization Tips

### 连接池配置 Connection Pool Configuration
工具已优化HTTP连接池配置，自动根据线程数调整连接池大小：
The tool has optimized HTTP connection pool configuration, automatically adjusting pool size based on thread count:
- `MaxIdleConns`: 等于线程数 / Equals thread count
- `MaxIdleConnsPerHost`: 等于线程数 / Equals thread count  
- `MaxConnsPerHost`: 等于线程数 / Equals thread count

### 最佳实践 Best Practices
1. **合理设置线程数**：建议从较小值开始（如10-20），逐步增加观察性能变化
   **Set appropriate thread count**: Start with smaller values (e.g., 10-20) and gradually increase while monitoring performance

2. **启用Keep-Alive**：默认启用，可显著提高性能，减少连接建立开销
   **Enable Keep-Alive**: Enabled by default, significantly improves performance by reducing connection overhead

3. **调整超时时间**：根据目标服务器响应速度调整timeout参数
   **Adjust timeout**: Set timeout parameter based on target server response speed

4. **监控资源使用**：高并发测试时注意监控客户端CPU和内存使用
   **Monitor resource usage**: Monitor client CPU and memory usage during high-concurrency tests

## 安全提示 Security Notes

### TLS证书验证 TLS Certificate Verification
- **生产环境**：建议保持 `skipVerify=false`，确保证书验证
  **Production**: Keep `skipVerify=false` to ensure certificate verification
  
- **测试环境**：可使用 `skipVerify=true` 跳过自签名证书验证
  **Testing**: Use `skipVerify=true` to skip verification for self-signed certificates

### 客户端证书认证 Client Certificate Authentication
使用双向TLS认证时，需要同时提供：
For mutual TLS authentication, provide all three files:
- 客户端证书（clientCert）
- 客户端私钥（clientKey）
- CA证书（caCert）

### 负载测试注意事项 Load Testing Considerations
- 确保有权对目标服务器进行压力测试
  Ensure you have permission to stress test the target server
- 避免对生产环境进行过大压力的测试
  Avoid excessive load testing on production environments
- 注意遵守目标服务的使用条款
  Respect the target service's terms of use



---
🌟 **如果您觉得这个工具有用，请给个Star！** ⭐️
*If you find this tool helpful, please consider giving it a star!* ⭐️

## TODO 待办

- [x] 支持结果输出为文件
- [x] 支持结果输出JSON
- [ ] 添加请求发送进度条显示
- [ ] 支持 CSV 格式输出

