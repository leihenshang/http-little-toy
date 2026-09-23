# http-little-toy

简单而强大的 HTTP 并发压测小工具，Go 语言编写，单二进制零依赖运行。如果觉得好用，请点个 Star ⭐️！
A simple yet powerful HTTP concurrency testing tool written in Go. Star ⭐️ if you find it useful!

## 特性 Key Features

- **并发压测** 多协程并发请求，支持 HTTP/1.1 / HTTP/2 / Concurrency testing with HTTP/1.1 & HTTP/2
- **灵活配置** 自定义 Header、Body、超时、TLS/mTLS / Rich CLI options: headers, body, timeout, TLS/mTLS
- **详细统计** RPS、传输速率、平均/最快/最慢响应时间、成功失败数 / Detailed stats: RPS, transfer rate, latency, success/failure
- **多种输出** raw / JSON / CSV，可保存到文件 / Multiple output formats, save to file
- **跨平台** Windows / Linux / macOS / Cross-platform

## 安装 Installation

```bash
# Go Install（推荐 Recommended）
go install github.com/leihenshang/http-little-toy@latest

# 或源码编译 Or build from source
git clone https://github.com/leihenshang/http-little-toy.git
cd http-little-toy && go build -o http-little-toy
```

## 快速开始 Quick Start

```bash
# GET 压测：10 并发，持续 30 秒 / 10 threads for 30 seconds
http-little-toy -u http://example.com -t 10 -d 30

# POST + 自定义 Header + JSON Body
http-little-toy -u http://api.example.com/users -m POST \
  -header "Content-Type: application/json" \
  -body '{"name":"test"}' -t 20 -d 60

# JSON 输出并保存到文件 / JSON output to file
http-little-toy -u http://example.com -format json -resFile result.json
```

常用参数速查（完整说明见 [doc/cli-reference.md](doc/cli-reference.md)）：

| 参数 | 说明 Description | 默认 Default |
|------|------------------|--------------|
| `-u` | 目标 URL（必填）Target URL (required) | - |
| `-m` | HTTP 方法 HTTP method | GET |
| `-t` | 并发线程数 Concurrent threads | 10 |
| `-d` | 持续时间(秒) Duration (seconds) | 10 |
| `-timeout` | 单请求超时(秒) Request timeout (seconds) | 10 |
| `-header` | 自定义 Header（可重复）Custom header (repeatable) | [] |
| `-body` | 请求体 Request body | "" |
| `-format` | 输出格式 raw/json/csv Output format | raw |
| `-h` / `-v` | 帮助 / 版本 Help / Version | - |

## 示例输出 Example Output

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

测试过程中按 `Ctrl+C` 可提前结束并输出当前统计 / Press `Ctrl+C` to stop early and print stats so far.

## 文档 Documentation

- [命令行参数完整说明 Full CLI Reference](doc/cli-reference.md)
- [跨平台编译 Cross-platform Build Guide](doc/build.md)
- [高级用法、性能优化与安全建议 Advanced Usage, Performance Tips & Security Notes](doc/advanced-usage.md)
- [test-server 本地测试服务 Local Test Server](test-server/README.md)

## 技术要求 Technical Requirements

- Go 1.22+，除 `golang.org/x/net/http2` 外仅使用标准库 / Standard library only besides `golang.org/x/net/http2`

## TODO

- [x] 结果输出为文件 Save result to file
- [x] 结果输出 JSON JSON output
- [ ] 请求发送进度条 Progress bar
- [ ] CSV 格式输出完善 Full CSV output support

---
🌟 **如果您觉得这个工具有用，请给个Star！** *If you find this tool helpful, please consider giving it a star!* ⭐️
