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
			name: "no errors",
			errs: nil,
			wantFn: func(s string) bool {
				return s == ""
			},
		},
		{
			name: "same error repeats",
			errs: []error{errors.New("boom"), errors.New("boom")},
			wantFn: func(s string) bool {
				return strings.Contains(s, "all failed requests share the error: boom")
			},
		},
		{
			name: "first and last differ",
			errs: []error{errors.New("first-err"), errors.New("mid"), errors.New("last-err")},
			wantFn: func(s string) bool {
				return strings.Contains(s, "first-err") && strings.Contains(s, "last-err") && !strings.Contains(s, "mid")
			},
		},
		{
			name: "nil error counted but no message",
			errs: []error{nil},
			wantFn: func(s string) bool {
				return s == ""
			},
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

func TestRenderLine(t *testing.T) {
	t.Run("half way", func(t *testing.T) {
		line := renderLine(5*time.Second, 10*time.Second, 472, 3, 5*1024)
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
		// 定宽填充（lineWidth + 1 个 \r），防止长行残留鬼影 / padded to fixed width
		if len(line) != lineWidth+1 {
			t.Errorf("len(line) = %d, want %d", len(line), lineWidth+1)
		}
	})

	t.Run("elapsed beyond total clamps to 100", func(t *testing.T) {
		line := renderLine(20*time.Second, 10*time.Second, 1, 0, 1)
		if !strings.Contains(line, "100%") || !strings.Contains(line, "10/10s") {
			t.Errorf("expected clamped 100%%, got %q", line)
		}
		if strings.Contains(line, ">") {
			t.Error("completed bar should not contain '>' marker")
		}
	})

	t.Run("zero total has no NaN", func(t *testing.T) {
		line := renderLine(0, 0, 0, 0, 0)
		if strings.Contains(line, "NaN") || strings.Contains(line, "%!") {
			t.Errorf("formatting broken: %q", line)
		}
	})
}

func TestRenderFinishesWithNewline(t *testing.T) {
	tracker := NewTracker()
	tracker.OnSuccess(128)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // ctx 已结束时 Render 应立刻输出 100% 并补换行退出

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		Render(ctx, tracker, &buf, 5*time.Second, time.Millisecond)
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
