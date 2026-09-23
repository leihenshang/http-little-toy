# 命令行参数完整说明 Full CLI Reference

> 返回 [README](../README.md) · Back to [README](../README.md)

## 基本参数 Basic Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-u` | 测试目标URL（必填）Target URL (required) | - | `-u http://example.com` |
| `-m` | HTTP请求方法 HTTP request method | GET | `-m POST` |
| `-t` | 并发线程数 Number of concurrent threads | 10 | `-t 50` |
| `-d` | 测试持续时间(秒) Test duration (seconds) | 10 | `-d 60` |
| `-timeout` | 单个请求超时时间(秒) Request timeout (seconds) | 10 | `-timeout 30` |
| `-header` | 自定义HTTP头部（可多次使用）Custom HTTP headers (can be used multiple times) | [] | `-header "Key: Value"` |
| `-body` | 请求体内容 Request body content | "" | `-body '{"data":"test"}'` |

## 连接与协议参数 Connection & Protocol Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-keepAlive` | 启用HTTP长连接（提高性能）Enable HTTP keep-alive (improves performance) | true | `-keepAlive=false` |
| `-compression` | 启用HTTP压缩传输 Enable HTTP compression | true | `-compression=false` |
| `-h2` | 使用HTTP/2协议 Use HTTP/2 protocol | false | `-h2=true` |
| `-allowRedirects` | 允许HTTP重定向 Allow HTTP redirects | true | `-allowRedirects=false` |

## TLS/SSL 安全参数 TLS/SSL Security Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-skipVerify` | 跳过TLS证书验证（不安全）Skip TLS certificate verification (insecure) | false | `-skipVerify=true` |
| `-clientCert` | 客户端证书文件路径 Client certificate file path | "" | `-clientCert cert.pem` |
| `-clientKey` | 客户端私钥文件路径 Client private key file path | "" | `-clientKey key.pem` |
| `-caCert` | CA证书文件路径 CA certificate file path | "" | `-caCert ca.pem` |

注意：`-clientCert`、`-clientKey`、`-caCert` 三者需同时提供，用于双向 TLS 认证。
Note: for mutual TLS authentication, all three flags must be provided together.

## 输出参数 Output Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-resFile` | 结果输出文件路径 Result output file path | "" | `-resFile result.txt` |
| `-format` | 输出格式（raw/json/csv）Output format (raw/json/csv) | raw | `-format json` |
| `-lang` | 输出语言（en/zh）Output language (en/zh) | en | `-lang zh` |

说明：`-format json/csv` 时不再逐行打印 raw 输出，统计结果在测试结束后统一输出。
Note: with `-format json/csv`, incremental raw printing is disabled; stats are emitted at the end.

## 帮助参数 Help Parameters

| 参数 | 说明 Description | 默认值 Default | 示例 Example |
|------|------------------|----------------|--------------|
| `-h` | 显示帮助信息 Show help information | false | `-h` |
| `-v` | 显示版本信息 Show version information | false | `-v` |

## 输出结果解读 Output Interpretation

工具提供全面的测试结果 / The tool provides comprehensive test results including:

- **成功/失败次数 Success/Failure Count**: 成功和失败的请求数量 / Number of successful and failed requests
- **吞吐量 Throughput**: 每秒请求数(RPS) / Requests per second (RPS)
- **传输速率 Transfer Rate**: 数据传输速度(KB/s) / Data transfer speed (KB/s)
- **响应时间 Response Time**: 平均、最小和最大请求时间 / Average, minimum, and maximum request times
- **总数据量 Total Data**: 测试期间传输的数据总量 / Amount of data transferred during the test

## 连接池行为 Connection Pool Behavior

HTTP 连接池根据 `-t`（线程数）自动调整 / Connection pool is sized automatically from `-t`:
- `MaxIdleConns`: 等于线程数 / Equals thread count
- `MaxIdleConnsPerHost`: 等于线程数 / Equals thread count
- `MaxConnsPerHost`: 等于线程数 / Equals thread count
