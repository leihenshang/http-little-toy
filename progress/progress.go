// Package progress 提供零第三方依赖的终端进度条：纯 \r 回车覆盖渲染，
// 兼容 Linux/macOS/Windows 控制台，不依赖 ANSI 转义与 cgo。
// 帧宽动态适配终端列宽（标准库 syscall + 构建标签探测，探测失败回退定宽），
// 保证任何一帧都不会超过终端宽度——超长行会触发软换行，从而打破 \r 覆盖模型。
// Package progress provides a zero-third-party-dependency terminal progress
// bar rendered with carriage-return overwrite, compatible with consoles on
// Linux/macOS/Windows without ANSI or cgo.
package progress

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// maxBarWidth 进度条主体最大格数。刻意使用纯 ASCII，规避中文在部分
	// Windows 终端下字符宽度不一致导致的覆盖残影。
	maxBarWidth = 30

	// minBarWidth 降级收缩时进度条的最小格数。
	minBarWidth = 8

	// fallbackWidth 探测不到终端列宽时的安全宽度（兼容 80 列终端）。
	// fallbackWidth is used when the terminal width cannot be detected.
	fallbackWidth = 78

	// DefaultInterval 默认刷新间隔。渲染由独立协程按 tick 驱动，
	// worker 热路径只做原子自增，绝不逐请求渲染。
	// DefaultInterval is the refresh tick; workers never render per-request.
	DefaultInterval = 200 * time.Millisecond
)

// Tracker 供 worker 热路径无锁累加的实时计数器，
// 并记录首次/末次错误，供测试结束后聚合输出（替代逐条错误日志）。
// Tracker holds lock-free counters for worker hot paths and records the
// first/last errors for an aggregated summary instead of per-request logs.
type Tracker struct {
	Success atomic.Int64
	Failed  atomic.Int64
	Bytes   atomic.Int64

	mu       sync.Mutex
	firstErr string
	lastErr  string
}

func NewTracker() *Tracker {
	return &Tracker{}
}

// OnSuccess 在单次请求成功后调用 / called after one successful request.
func (t *Tracker) OnSuccess(bytes int) {
	t.Success.Add(1)
	t.Bytes.Add(int64(bytes))
}

// OnFailure 在单次请求失败后调用；err 允许为 nil（如空响应被判失败）。
// OnFailure is called after one failed request; err may be nil.
func (t *Tracker) OnFailure(err error) {
	t.Failed.Add(1)
	if err == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.firstErr == "" {
		t.firstErr = err.Error()
	}
	t.lastErr = err.Error()
}

// ErrorSummary 返回一行错误摘要；无错误时返回空字符串。
// ErrorSummary returns a one-line summary of failures, "" when none.
func (t *Tracker) ErrorSummary() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch {
	case t.firstErr == "":
		return ""
	case t.firstErr == t.lastErr:
		return fmt.Sprintf("all failed requests share the error: %s", t.firstErr)
	default:
		return fmt.Sprintf("first error: %s | last error: %s", t.firstErr, t.lastErr)
	}
}

// Width 返回 f 关联终端的列宽；非终端、不支持的平台或探测失败返回 0。
// 平台相关实现见 width_unix.go / width_windows.go / width_other.go。
// Width returns the terminal column count of f, or 0 when unknown.
func Width(f *os.File) int {
	return terminalWidth(f)
}

// ShouldRender 决定是否显示进度条：用户开关开启、输出格式为 raw
// （json/csv 需要干净的 stdout）、且 stdout 为交互终端。
// ShouldRender: enabled && format==raw && stdout is a character device
// (consoles on Linux/macOS/Windows qualify; pipes/files do not).
func ShouldRender(enabled bool, format string, out *os.File) bool {
	if !enabled || format != "raw" || out == nil {
		return false
	}
	info, err := out.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// Render 按 interval 周期性刷新进度条，ctx 结束后输出 100% 并补一个
// 换行，避免 shell 提示符与进度行粘连。
// widthFn 每次 tick 重新探测终端列宽（跟随窗口缩放）；返回 0 或为 nil 时
// 回退 fallbackWidth。
// Render refreshes the bar on each tick; when ctx is done it draws 100%
// and emits a final newline so the shell prompt stays clean. widthFn is
// re-sampled every tick so the bar follows terminal resize.
func Render(ctx context.Context, t *Tracker, out io.Writer, total time.Duration, interval time.Duration, widthFn func() int) {
	if interval <= 0 {
		interval = DefaultInterval
	}
	start := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	detected := 0
	if widthFn != nil {
		detected = widthFn()
	}
	effective := func() int {
		if detected > 0 {
			return detected
		}
		return fallbackWidth
	}

	draw := func(elapsed time.Duration) {
		fmt.Fprint(out, renderLine(elapsed, total, t.Success.Load(), t.Failed.Load(), t.Bytes.Load(), effective()))
	}

	for {
		select {
		case <-ticker.C:
			if widthFn != nil {
				detected = widthFn()
			}
			draw(time.Since(start))
		case <-ctx.Done():
			draw(total)
			fmt.Fprintln(out)
			return
		}
	}
}

// frameCounts 是一帧的可格式化数据；frameSpec 控制渲染降级程度。
type frameCounts struct {
	frac    float64
	elapsed time.Duration
	total   time.Duration
	success int64
	failed  int64
	rps     float64
	rateKB  float64
}

type frameSpec struct {
	barW     int
	withRate bool // 是否显示 RPS 与传输速率段
	abbrev   bool // 是否缩写大数（155311 → 155.3k）
}

// renderLine 为纯函数，便于测试。完整样式：
// [==============>               ]  48%  5/10s OK:472 ERR:3 RPS:94.4 4.7KB/s
//
// 超长时按序降级（② 静态限长兜底）：
//  1. 收缩进度条格数（信息全保留）
//  2. 去掉速率段（RPS/KB/s）
//  3. 大数字缩写（k/M）
//  4. 仍超长则硬截断——保证绝不触发软换行
func renderLine(elapsed, total time.Duration, success, failed, bytesRead int64, maxWidth int) string {
	if elapsed > total {
		elapsed = total
	}
	var frac float64
	if total > 0 {
		frac = float64(elapsed) / float64(total)
	}
	if frac > 1 {
		frac = 1
	}

	var rps, rateKB float64
	if sec := elapsed.Seconds(); sec > 0 {
		// RPS = 总请求速率（成功+失败）：目标变慢/宕机时仍能反映实际发送压力
		rps = float64(success+failed) / sec
		rateKB = float64(bytesRead) / sec / 1024
	}
	c := frameCounts{frac: frac, elapsed: elapsed, total: total, success: success, failed: failed, rps: rps, rateKB: rateKB}

	if maxWidth <= 0 {
		maxWidth = fallbackWidth
	}

	// barFit 计算在不丢信息前提下进度条最多能占多少格。
	barFit := func(withRate, abbrev bool) int {
		overhead := len(c.text(frameSpec{barW: 0, withRate: withRate, abbrev: abbrev}))
		w := maxWidth - overhead
		if w > maxBarWidth {
			w = maxBarWidth
		}
		if w < minBarWidth {
			w = minBarWidth
		}
		return w
	}

	specs := []frameSpec{
		{barW: maxBarWidth, withRate: true, abbrev: false},           // 完整
		{barW: barFit(true, false), withRate: true, abbrev: false},   // 收缩进度条
		{barW: barFit(false, false), withRate: false, abbrev: false}, // 丢速率段
		{barW: barFit(false, true), withRate: false, abbrev: true},   // 缩写大数
		{barW: minBarWidth, withRate: false, abbrev: true},           // 最小帧
	}
	for _, s := range specs {
		if text := c.text(s); len(text) <= maxWidth {
			return "\r" + padTo(text, maxWidth)
		}
	}
	text := c.text(specs[len(specs)-1])
	if len(text) > maxWidth {
		text = text[:maxWidth] // 纯 ASCII 内容，按字节截断安全
	}
	return "\r" + text
}

// text 按 spec 组装一帧内容（不含 \r 与补位空格）。
func (c frameCounts) text(s frameSpec) string {
	filled := int(c.frac * float64(s.barW))
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", s.barW-filled)
	if filled > 0 && filled < s.barW {
		bar = bar[:filled-1] + ">" + bar[filled:]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %3.0f%% %3.0f/%ds OK:%s ERR:%s",
		bar, c.frac*100, c.elapsed.Seconds(), int(c.total.Seconds()),
		abbrevInt(c.success, s.abbrev), abbrevInt(c.failed, s.abbrev))
	if s.withRate {
		fmt.Fprintf(&b, " RPS:%s %.1fKB/s", abbrevFloat(c.rps, s.abbrev), c.rateKB)
	}
	return b.String()
}

// abbrevInt 缩写大整数：≥100k 时转 k/M 后缀，如 155311 → 155.3k。
func abbrevInt(n int64, abbrev bool) string {
	if !abbrev {
		return strconv.FormatInt(n, 10)
	}
	return abbrevMagnitude(float64(n))
}

// abbrevFloat 缩写大浮点数，规则同 abbrevInt。
func abbrevFloat(v float64, abbrev bool) string {
	if !abbrev {
		return fmt.Sprintf("%.1f", v)
	}
	return abbrevMagnitude(v)
}

func abbrevMagnitude(v float64) string {
	switch {
	case v >= 1_000_000:
		return trimZero(fmt.Sprintf("%.1fM", v/1e6))
	case v >= 100_000:
		return trimZero(fmt.Sprintf("%.1fk", v/1e3))
	case v >= 1000:
		return trimZero(fmt.Sprintf("%.1fk", v/1e3))
	default:
		return trimZero(fmt.Sprintf("%.1f", v))
	}
}

// trimZero 去掉多余的 ".0" 小数尾，缩短宽度。
func trimZero(s string) string {
	return strings.Replace(s, ".0", "", 1)
}

// padTo 用空格补齐到 width 列，覆盖上一帧残留，防止鬼影。
func padTo(text string, width int) string {
	if pad := width - len(text); pad > 0 {
		return text + strings.Repeat(" ", pad)
	}
	return text
}
