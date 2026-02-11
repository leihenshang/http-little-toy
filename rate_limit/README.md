基于当前代码基础，我为您设计一个负载测试实现方案：

## 核心设计思路

1. **添加QPS控制参数**
   - 新增 `-qps` 参数指定目标QPS
   - 默认值设为0表示不限制（兼容现有行为）

2. **实现令牌桶限流器**
   - 创建独立的rate_limiter包
   - 支持动态QPS调整
   - 提供阻塞和非阻塞两种获取模式

3. **改造请求发送逻辑**
   - 在doReq调用前加入限流检查
   - 支持突发模式和稳定模式切换
   - 保持现有统计逻辑不变

## 具体实现要点

4. **命令行接口扩展**
   ```
   -qps int     Target requests per second (default 0, unlimited)
   -burst int   Burst size for token bucket (default 10)
   -mode string Request mode: stable|burst (default "burst")
   ```

5. **关键代码结构调整**
   - 将限流逻辑封装为独立组件
   - 在主循环中集成限流检查
   - 保持向后兼容性

6. **性能监控增强**
   - 记录实际达到的QPS
   - 统计限流等待时间
   - 提供实时QPS监控

这个方案能够在不影响现有功能的前提下，为工具增加专业的负载测试能力。