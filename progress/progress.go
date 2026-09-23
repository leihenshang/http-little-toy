// Package progress 提供零依赖的终端进度条：纯 \r 回车覆盖渲染，
// 兼容 Linux/macOS/Windows 控制台，不依赖 ANSI 转义与 cgo。
// Package progress provides a zero-dependency terminal progress bar rendered
// with carriage-return overwrite, compatible with Linux/macOS/Windows consoles.
package progress

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// barWidth 进度条主体宽度。刻意使用纯 ASCII，规避中文在部分 Windows
	// 终端下字符宽度不一致导致的覆盖残影。
	// barWidth is the gauge body width; ASCII-only to avoid CJK width issues.
	barWidth = 30

	// lineWidth 渲染行总宽（不含 \r），定宽填充覆盖上一行防止残影。
	// lineWidth is the padded line width (excluding \r) to avoid leftovers.
	lineWidth = 78

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
// Render refreshes the bar on each tick; when ctx is done it draws 100%
// and emits a final newline so the shell prompt stays clean.
func Render(ctx context.Context, t *Tracker, out io.Writer, total time.Duration, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultInterval
	}
	start := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprint(out, renderLine(time.Since(start), total, t.Success.Load(), t.Failed.Load(), t.Bytes.Load()))
		case <-ctx.Done():
			fmt.Fprint(out, renderLine(total, total, t.Success.Load(), t.Failed.Load(), t.Bytes.Load()))
			fmt.Fprintln(out)
			return
		}
	}
}

// renderLine 为纯函数，便于测试。样式：
// [==============>               ]  48%  5/10s OK:472 ERR:3 RPS:94.4 4.7KB/s
func renderLine(elapsed, total time.Duration, success, failed, bytes int64) string {
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

	filled := int(frac * barWidth)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
	if filled > 0 && filled < barWidth {
		bar = bar[:filled-1] + ">" + bar[filled:]
	}

	var rps, rateKB float64
	if sec := elapsed.Seconds(); sec > 0 {
		rps = float64(success) / sec
		rateKB = float64(bytes) / sec / 1024
	}

	text := fmt.Sprintf("[%s] %3.0f%% %3.0f/%ds OK:%d ERR:%d RPS:%.1f %.1fKB/s",
		bar, frac*100, elapsed.Seconds(), int(total.Seconds()), success, failed, rps, rateKB)
	if pad := lineWidth - len(text); pad > 0 {
		text += strings.Repeat(" ", pad)
	}
	return "\r" + text
}
