# 高级用法、性能优化与安全建议 Advanced Usage, Performance Tips & Security Notes

> 返回 [README](../README.md) · Back to [README](../README.md)

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
# 复杂API测试（JSON负载）Complex API test with JSON payload
./http-little-toy -u http://localhost:2025/api/users/list -m POST \
  -header "token: itoken_1754640941655527771" \
  -header "Content-Type: application/json" \
  -body '{
    "pageSize": 100,
    "page": 1
  }' -d 60 -t 30
```

### HTTPS 证书认证测试 HTTPS with Certificate Authentication

```bash
# 双向TLS认证（客户端证书 + 私钥 + CA证书）Mutual TLS: client cert + key + CA cert
./http-little-toy -u https://secure-api.com \
  -clientCert client.crt \
  -clientKey client.key \
  -caCert ca.crt \
  -skipVerify=false \
  -d 30 -t 20
```

### 输出控制 Output Control

```bash
# 输出JSON格式结果 Output results in JSON format
http-little-toy -u http://example.com -t 10 -d 30 -format json

# 将结果保存到文件 Save results to file
http-little-toy -u http://example.com -t 10 -d 30 -resFile result.txt

# JSON格式并保存到文件 JSON format and save to file
http-little-toy -u http://example.com -t 10 -d 30 -format json -resFile result.json

# 中文输出 Chinese output
http-little-toy -u http://example.com -t 10 -d 30 -lang zh
```

## 性能优化建议 Performance Optimization Tips

### 连接池配置 Connection Pool Configuration

工具已优化 HTTP 连接池配置，自动根据线程数调整连接池大小 / The connection pool is auto-sized from thread count:
- `MaxIdleConns` / `MaxIdleConnsPerHost` / `MaxConnsPerHost`：均等于线程数 / all equal to thread count

### 最佳实践 Best Practices

1. **合理设置线程数**：建议从较小值开始（如 10-20），逐步增加观察性能变化
   **Set appropriate thread count**: Start small (e.g. 10-20) and gradually increase while monitoring
2. **启用 Keep-Alive**：默认启用，可显著减少连接建立开销
   **Enable Keep-Alive**: Enabled by default, greatly reduces connection overhead
3. **调整超时时间**：根据目标服务器响应速度调整 `timeout` 参数
   **Adjust timeout**: Set `timeout` based on target server response speed
4. **监控资源使用**：高并发测试时注意监控客户端 CPU 和内存使用
   **Monitor resource usage**: Watch client CPU/memory during high-concurrency tests

## 安全提示 Security Notes

### TLS 证书验证 TLS Certificate Verification

- **生产环境 Production**：建议保持 `skipVerify=false`，确保证书验证 / Keep `skipVerify=false` to ensure certificate verification
- **测试环境 Testing**：可使用 `skipVerify=true` 跳过自签名证书验证 / Use `skipVerify=true` for self-signed certificates

### 客户端证书认证 Client Certificate Authentication

使用双向 TLS 认证时，需要同时提供客户端证书（`clientCert`）、客户端私钥（`clientKey`）与 CA 证书（`caCert`）。
For mutual TLS, provide all three: client cert, client key, and CA cert.

### 负载测试注意事项 Load Testing Considerations

- 确保有权对目标服务器进行压力测试 / Ensure you have permission to stress test the target
- 避免对生产环境进行过大压力的测试 / Avoid excessive load testing on production
- 注意遵守目标服务的使用条款 / Respect the target service's terms of use
