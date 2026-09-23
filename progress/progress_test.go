package progress

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTrackerConcurrentCounters(t *testing.T) {
	tracker := NewTracker()

	var wg sync.WaitGroup
	const workers, perWorker = 8, 500
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				if j%2 == 0 {
					tracker.OnSuccess(10)
				} else {
					tracker.OnFailure(fmt.Errorf("worker %d err %d", id, j))
				}
			}
		}(i)
	}
	wg.Wait()

	if got := tracker.Success.Load(); got != workers*perWorker/2 {
		t.Errorf("Success = %d, want %d", got, workers*perWorker/2)
	}
	if got := tracker.Failed.Load(); got != workers*perWorker/2 {
		t.Errorf("Failed = %d, want %d", got, workers*perWorker/2)
	}
	if got := tracker.Bytes.Load(); got != int64(workers*perWorker/2)*10 {
		t.Errorf("Bytes = %d, want %d", got, int64(workers*perWorker/2)*10)
	}
}

func TestTrackerErrorSummary(t *testing.T) {
	tests := []struct {
		name   string
		errs   []error
		wantFn func(s string) bool
	}{
		{
			name:   "no errors",
			errs:   nil,
			wantFn: func(s string) bool { return s == "" },
		},
		{
			name:   "same error repeats",
			errs:   []error{errors.New("boom"), errors.New("boom")},
			wantFn: func(s string) bool { return strings.Contains(s, "all failed requests share the error: boom") },
		},
		{
			name: "first and last differ",
			errs: []error{errors.New("first-err"), errors.New("mid"), errors.New("last-err")},
			wantFn: func(s string) bool {
				return strings.Contains(s, "first-err") && strings.Contains(s, "last-err") && !strings.Contains(s, "mid")
			},
		},
		{
			name:   "nil error counted but no message",
			errs:   []error{nil},
			wantFn: func(s string) bool { return s == "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker()
			for _, err := range tt.errs {
				tracker.OnFailure(err)
			}
			if !tt.wantFn(tracker.ErrorSummary()) {
				t.Errorf("ErrorSummary() = %q, unexpected", tracker.ErrorSummary())
			}
		})
	}
}

func TestShouldRender(t *testing.T) {
	// 普通文件不是字符设备，进度条应自动关闭 / a regular file is no TTY
	tmpPath := filepath.Join(t.TempDir(), "out.txt")
	f, err := os.Create(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tests := []struct {
		name     string
		enabled  bool
		format   string
		out      *os.File
		expected bool
	}{
		{"switch off", false, "raw", os.Stdout, false},
		{"json format", true, "json", os.Stdout, false},
		{"csv format", true, "csv", os.Stdout, false},
		{"nil writer", true, "raw", nil, false},
		{"regular file not tty", true, "raw", f, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldRender(tt.enabled, tt.format, tt.out); got != tt.expected {
				t.Errorf("ShouldRender(%v, %q) = %v, want %v", tt.enabled, tt.format, got, tt.expected)
			}
		})
	}
}

// 高吞吐场景：数字位数把完整帧撑到 ~84 列，窄终端上必须降级而非换行
func highThroughputLine(maxWidth int) string {
	return renderLine(9600*time.Millisecond, 10*time.Second, 149387, 3, 3_200_000, maxWidth)
}

func TestRenderLine(t *testing.T) {
	t.Run("half way", func(t *testing.T) {
		line := renderLine(5*time.Second, 10*time.Second, 472, 3, 5*1024, fallbackWidth)
		if !strings.HasPrefix(line, "\r") {
			t.Error("line must start with \\r for carriage-return overwrite")
		}
		if !strings.Contains(line, "50%") || !strings.Contains(line, "5/10s") {
			t.Errorf("unexpected progress text: %q", line)
		}
		if !strings.Contains(line, ">") {
			t.Error("in-progress bar should contain '>' head marker")
		}
		if !strings.Contains(line, "OK:472") || !strings.Contains(line, "ERR:3") {
			t.Errorf("counters missing: %q", line)
		}
		// RPS 按总请求数（成功+失败）计算：(472+3)/5s = 95.0
		if !strings.Contains(line, "RPS:95.0") {
			t.Errorf("RPS should count success+failed: got %q, want RPS:95.0", line)
		}
		// 短帧补齐到 maxWidth 列（fallbackWidth + 1 个 \r），防止长帧残留鬼影
		if len(line) != fallbackWidth+1 {
			t.Errorf("len(line) = %d, want %d", len(line), fallbackWidth+1)
		}
	})

	t.Run("elapsed beyond total clamps to 100", func(t *testing.T) {
		line := renderLine(20*time.Second, 10*time.Second, 1, 0, 1, fallbackWidth)
		if !strings.Contains(line, "100%") || !strings.Contains(line, "10/10s") {
			t.Errorf("expected clamped 100%%, got %q", line)
		}
		if strings.Contains(line, ">") {
			t.Error("completed bar should not contain '>' marker")
		}
	})

	t.Run("zero total has no NaN", func(t *testing.T) {
		line := renderLine(0, 0, 0, 0, 0, fallbackWidth)
		if strings.Contains(line, "NaN") || strings.Contains(line, "%!") {
			t.Errorf("formatting broken: %q", line)
		}
	})

	// 目标宕机、全部失败时 RPS 仍应反映发送速率（旧实现只统计成功，显示 0.0）
	t.Run("all failed still shows send rate", func(t *testing.T) {
		line := renderLine(10*time.Second, 10*time.Second, 0, 500, 0, fallbackWidth)
		if !strings.Contains(line, "RPS:50.0") || !strings.Contains(line, "ERR:500") {
			t.Errorf("RPS must include failures: %q", line)
		}
	})
}

// TestRenderLineNeverExceedsWidth 核心属性（修复"帧宽超终端导致软换行、
// \r 覆盖失效、帧首尾相连"的问题）：任何一帧的可视长度都必须 ≤ 终端列宽。
func TestRenderLineNeverExceedsWidth(t *testing.T) {
	for width := 12; width <= 120; width++ {
		line := highThroughputLine(width)
		text := strings.TrimPrefix(line, "\r")
		if len(text) > width {
			t.Errorf("maxWidth=%d: rendered %d cols, would soft-wrap: %q", width, len(text), text)
		}
		if strings.ContainsAny(text, "\r\n") {
			t.Errorf("maxWidth=%d: embedded newline in frame: %q", width, text)
		}
	}
}

// TestRenderLineDegradesInOrder 验证降级顺序：先收缩 bar（信息保留），
// 再丢速率段，再缩写大数。
func TestRenderLineDegradesInOrder(t *testing.T) {
	t.Run("wide keeps full info", func(t *testing.T) {
		line := highThroughputLine(100)
		if !strings.Contains(line, "RPS:") || !strings.Contains(line, "OK:149387") {
			t.Errorf("wide terminal should keep rate and exact counts: %q", line)
		}
	})
	t.Run("70 cols shrinks bar but keeps info", func(t *testing.T) {
		line := highThroughputLine(70)
		if !strings.Contains(line, "RPS:") || !strings.Contains(line, "OK:149387") {
			t.Errorf("bar should shrink before dropping information: %q", line)
		}
		if len(strings.TrimSpace(line)) < 70 {
			// bar 被收缩的证据：整帧（除补齐）短于 70 说明不是靠截断
			t.Logf("frame shortened as expected: %q", line)
		}
	})
	t.Run("50 cols drops rate", func(t *testing.T) {
		line := highThroughputLine(50)
		if strings.Contains(line, "KB/s") {
			t.Errorf("narrow frame should drop transfer rate: %q", line)
		}
		if !strings.Contains(line, "OK:") {
			t.Errorf("success counter must survive: %q", line)
		}
	})
	t.Run("very narrow abbreviates or truncates", func(t *testing.T) {
		line := highThroughputLine(30)
		if !strings.Contains(line, "OK:") {
			t.Errorf("frame at 30 cols lost OK counter: %q", line)
		}
		if len(strings.TrimPrefix(line, "\r")) > 30 {
			t.Errorf("frame exceeds width at 30 cols: %q", line)
		}
	})
}

func TestAbbrevInt(t *testing.T) {
	tests := []struct {
		in     int64
		abbrev bool
		want   string
	}{
		{472, false, "472"},
		{149387, false, "149387"},
		{472, true, "472"},
		{1500, true, "1.5k"},
		{95000, true, "95k"},
		{149387, true, "149.4k"},
		{1_500_000, true, "1.5M"},
	}
	for _, tt := range tests {
		if got := abbrevInt(tt.in, tt.abbrev); got != tt.want {
			t.Errorf("abbrevInt(%d, %v) = %q, want %q", tt.in, tt.abbrev, got, tt.want)
		}
	}
}

func TestWidthOnNonTerminalIsZero(t *testing.T) {
	// 普通文件不是终端，探测应安全失败返回 0 / regular file has no width
	f, err := os.Create(filepath.Join(t.TempDir(), "w.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got := Width(f); got != 0 {
		t.Errorf("Width(regular file) = %d, want 0", got)
	}
	if got := Width(nil); got != 0 {
		t.Errorf("Width(nil) = %d, want 0", got)
	}
}

func TestRenderFinishesWithNewline(t *testing.T) {
	tracker := NewTracker()
	tracker.OnSuccess(128)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // ctx 已结束时 Render 应立刻输出 100% 并补换行退出

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		Render(ctx, tracker, &buf, 5*time.Second, time.Millisecond, func() int { return 80 })
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Render did not return after ctx cancel")
	}

	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output must end with newline to keep shell prompt clean, got %q", out)
	}
	if !strings.Contains(out, "100%") {
		t.Errorf("final frame should show 100%%, got %q", out)
	}
}
