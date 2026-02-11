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

## 命令行参数 Command Line Options

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-u` | 测试目标URL Target URL to test | 必填 Required | `-u http://example.com` |
| `-m` | HTTP方法 HTTP method | GET | `-m POST` |
| `-t` | 并发线程数 Number of concurrent threads | 10 | `-t 50` |
| `-d` | 测试持续时间(秒) Test duration (seconds) | 10 | `-d 60` |
| `-header` | 自定义HTTP头部 Custom HTTP headers | [] | `-header "Key: Value"` |
| `-body` | 请求体内容 Request body | "" | `-body '{"data":"test"}'` |
| `-keepAlive` | 启用HTTP长连接 Enable HTTP keep-alive | true | `-keepAlive=false` |
| `-compression` | 启用压缩 Enable compression | true | `-compression=false` |
| `-timeout` | 请求超时时间(秒) Request timeout (seconds) | 5 | `-timeout 10` |
| `-h2` | 使用HTTP/2 Use HTTP/2 | false | `-h2=true` |
| `-skipVerify` | 跳过TLS证书验证 Skip TLS certificate verification | false | `-skipVerify=true` |
| `-allowRedirects` | 允许HTTP重定向 Allow HTTP redirects | true | `-allowRedirects=false` |
| `-clientCert` | 客户端证书文件 Client certificate file | "" | `-clientCert cert.pem` |
| `-clientKey` | 客户端密钥文件 Client key file | "" | `-clientKey key.pem` |
| `-caCert` | CA证书文件 CA certificate file | "" | `-caCert ca.pem` |
| `-h` | 显示帮助信息 Show help | false | `-h` |
| `-v` | 显示版本信息 Show version | false | `-v` |

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



---
🌟 **如果您觉得这个工具有用，请给个Star！** ⭐️
*If you find this tool helpful, please consider giving it a star!* ⭐️