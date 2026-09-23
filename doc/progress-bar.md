# 进度条实现原理与设计 Progress Bar Design & Internals

> 返回 [README](../README.md) · Back to [README](../README.md)

本文档说明 http-little-toy 进度条的实现逻辑与设计取舍。代码位于 [progress/progress.go](../progress/progress.go)，接入点在 [main.go](../main.go)。
This document explains how the progress bar works and the design trade-offs. Code lives in [progress/progress.go](../progress/progress.go); integration points are in [main.go](../main.go).

## 设计目标 Design Goals

1. **零依赖**：只用 Go 标准库，维持项目「除 `golang.org/x/net` 外无第三方依赖」与 `CGO_ENABLED=0` 交叉编译的现状。
   Zero third-party dependency; keeps the CGO-free cross-compile story intact.
2. **跨平台**：Linux / macOS / Windows 的 cmd、PowerShell 全部可用，不写任何平台专属 syscall。
   Works on every platform without a single platform-specific syscall.
3. **不伤害压测精度**：渲染开销不得干扰 worker 热路径。
   Rendering must never slow down or serialize the request workers.
4. **机器输出干净**：JSON/CSV 输出与管道重定向场景下自动关闭。
   Auto-disabled when stdout is not a terminal or output is machine-readable.

## 核心原理：`\r` 回车覆盖 Carriage-Return Overwrite

终端里的 `stdout` 只是字节流，所谓「动态刷新」其实是**在同一行反复重写**：

- `\r`（回车符）把光标移回**行首**但不换行；
- 之后写入的新内容从行首开始**逐格覆盖**旧字符；
- 人眼看到的「进度条在动」，是每 200ms 一次的整行重绘。

`renderLine` 的返回值永远是 `"\r" + 新行内容`，这就是显示层的全部核心。

**为什么不用 ANSI 转义序列**（`\033[K`、光标移动等）：ANSI 在 Windows 传统控制台上需要先开启 VT 模式（`SetConsoleMode`），要么引入 `golang.org/x/term`、要么写 `//go:build windows` 的平台分支。而 `\r` 从 DOS 时代起就是全平台通用能力——这是方案「纯 `\r` 覆盖」跨平台成本最低的根基。

### 定宽填充防「鬼影」Fixed-width Padding

若新一帧比上一帧短，旧帧尾部会残留（例如 `OK:9998` 变 `OK:42` 后剩 `8`）。因此每行统一补齐到 78 列（`lineWidth`），任何一帧都能完整覆盖上一帧：

```go
if pad := lineWidth - len(text); pad > 0 {
    text += strings.Repeat(" ", pad)
}
return "\r" + text
```

渲染效果示例（实际为单行滚动覆盖）：

```
[==========>                   ]  36%  1/3s OK:9865 ERR:0 RPS:8972.4 1813.5KB/s
[=====================>        ]  73%  2/3s OK:19925 ERR:0 RPS:9056.8 1830.8KB/s
[==============================] 100%  3/3s OK:27041 ERR:0 RPS:9013.7 1822.1KB/s
```

### 为什么进度条本体是纯 ASCII ASCII-only Bar

`\r` 覆盖的隐含假设是「写 N 个字符 = 占 N 列」。中文在 CJK 终端占 2 列但只算 1 个字符（Go 的 `len()` 还按字节计），宽度错位就会出现覆盖不全的残影。因此 `OK:` / `ERR:` / `RPS:` 标签全部用 ASCII，中文只出现在静止输出行。

### 收尾换行 Trailing Newline

测试结束（含 `Ctrl+C` 提前终止）时，`Render` 输出 100% 帧后补一个 `\n`，让 shell 提示符落在进度行下方，避免粘连。

## 架构：统计与渲染解耦 Producers / Renderer Separation

压测 worker 每秒可能完成数千个请求，数据生产者与消费者的速度差可达 4 个数量级：

```
N 个 worker（每秒数千次写） ──atomic.Add──▶ Tracker ──Load──▶ 渲染协程（每秒 5 次读）
```

### Tracker：热路径原子计数

[Tracker](../progress/progress.go) 用 `atomic.Int64` 维护三个实时计数器，worker 侧的 `OnSuccess`/`OnFailure` 是无锁、纳秒级的 `Add()`，不会阻塞请求循环：

```go
type Tracker struct {
    Success atomic.Int64
    Failed  atomic.Int64
    Bytes   atomic.Int64
    // 首错/末错用互斥锁记录，仅发生在错误路径（非热路径），内存恒定
    mu       sync.Mutex
    firstErr string
    lastErr  string
}
```

**为什么渲染绝不跟着请求走**：若每个请求完成时直接打印，`fmt.Fprint` 内部的 stdout 锁会把所有 worker 串行化，压测吞吐与延迟统计直接失真。渲染只在独立协程的 ticker 里发生——典型的**多生产者-单消费者 + 降采样**模式：计数不丢（原子性保证），画面按需节流（默认 200ms）。

### 错误聚合输出 Aggregated Error Summary

worker 不再逐条 `log.Printf` 打印失败（高并发下会刷屏、争用 log 锁、打乱进度条），而是由 `Tracker.OnFailure` 记录**首次与末次**错误，测试结束后输出一行摘要（`ErrorSummary`），同时写入 `-resFile`：

```
all failed requests share the error: Get "http://127.0.0.1:9092": dial tcp: connection refused
```

## 进度语义：为什么按时间而不是请求数 Time-based Progress

本工具是「`-t` 并发跑 `-d` 秒」模型，**总请求数没有上界**，无法计算「已完成/总量」。而流逝时间 / 设定时长是天然的 0→100%：

```go
frac = float64(elapsed) / float64(total)   // clamp 到 [0, 1]
```

请求数（OK / ERR / RPS / KB/s）作为辅助指标并排显示。`elapsed` 超过 `total` 时钳制为 100%，防止 ticker 抖动画出超界进度。

## 环境探测：字符设备判断「真终端」Char-device TTY Detection

`ShouldRender(enabled, format, out)` 依次检查三个条件，任何一条不满足即静默关闭进度条：

1. 用户开关 `-progress` 为 true；
2. `-format` 为 `raw`（`json`/`csv` 需要干净的 stdout 供机器解析）；
3. `out.Stat()` 的 `Mode & os.ModeCharDevice != 0`。

第 3 条是跨平台 isatty 的「够用近似」：控制台（Linux tty、Windows conhost）在 Go `os` 层呈现为**字符设备**；而 `> file`、`| jq` 重定向后 stdout 变成普通文件/管道，不再是字符设备。三平台行为一致，无需 syscall。

已知边界：重定向到 `/dev/null` 时因其同为字符设备，进度条仍会输出——无害且几乎不产生开销。

## main.go 接入点 Integration

```go
// 1. 参数注册（initParameters）
flag.BoolVar(&progressOn, "progress", true, "show progress bar (only in raw format on an interactive terminal).")

// 2. 创建 tracker、按条件启动渲染协程，与测试共用同一个 ctx
tracker := progress.NewTracker()
if progress.ShouldRender(progressOn, format, os.Stdout) {
    go progress.Render(ctx, tracker, os.Stdout, time.Duration(toyReq.Duration)*time.Second, progress.DefaultInterval)
}

// 3. worker 热路径：成功/失败只做原子计数
tracker.OnSuccess(size)   // 成功
tracker.OnFailure(err)    // 失败（不再逐条打日志）

// 4. 结束后输出错误摘要（追加进 Res，随 -resFile 一起落盘）
if summary := tracker.ErrorSummary(); summary != "" {
    allAggregate.Res = append(allAggregate.Res, summary)
    printLByFormat(format, summary)
}
```

`ctx` 是测试时长控制的 `context.WithTimeout`，`Ctrl+C` 信号处理里也会 `cancel()`——因此渲染协程的退出与测试终止**共享同一条生命周期链路**，不需要额外的停止信号。

## 测试 Testing

单元测试见 [progress/progress_test.go](../progress/progress_test.go)（`go test -race ./progress/`）：

| 用例 | 验证点 |
|------|--------|
| `TestTrackerConcurrentCounters` | 8 协程并发计数不丢（`-race`） |
| `TestTrackerErrorSummary` | 无错/同错/首末异错/nil 错误四种摘要形态 |
| `TestShouldRender` | 开关、json/csv、nil writer、普通文件（非 tty）均正确关闭 |
| `TestRenderLine` | 50% 帧含 `>` 头标、超界钳制 100%、定宽 78 补齐、无 NaN |
| `TestRenderFinishesWithNewline` | ctx 结束后立即输出 100% 并以 `\n` 收尾 |

## 一页总结 Summary

| 问题 | 答案 |
|------|------|
| 动态刷新怎么实现？ | `\r` 回行首 + 整行覆盖重绘，全平台通用、零依赖 |
| 残留鬼影怎么消除？ | 定宽 78 列填充 + 纯 ASCII 内容 |
| 会不会拖慢压测？ | 不会：worker 只做原子 Add，渲染由独立协程 200ms 节流驱动 |
| 进度按什么算？ | 时间（elapsed/duration），请求数无上界 |
| Windows 兼容？ | 天然兼容，不用 ANSI、不用 syscall、不用 cgo |
| 重定向/JSON 场景？ | `os.ModeCharDevice` 探测 + format==raw 检查，自动静默 |
