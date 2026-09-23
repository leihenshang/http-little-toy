package data

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRequestStats_PrintStats(t *testing.T) {
	tests := []struct {
		name  string
		stats *RequestStats
	}{
		{
			name: "Empty stats",
			stats: &RequestStats{
				Format: "raw",
				Url:    "http://example.com",
			},
		},
		{
			name: "Stats with data",
			stats: &RequestStats{
				Format:     "raw",
				Url:        "http://example.com",
				RespSize:   1024,
				Duration:   time.Second * 5,
				MinReqTime: time.Millisecond * 100,
				MaxReqTime: time.Millisecond * 500,
				ErrNum:     1,
				SuccessNum: 10,
				RespNum:    10,
			},
		},
		{
			name: "JSON format stats",
			stats: &RequestStats{
				Format:     "json",
				Url:        "http://example.com",
				RespSize:   1024,
				Duration:   time.Second * 5,
				MinReqTime: time.Millisecond * 100,
				MaxReqTime: time.Millisecond * 500,
				ErrNum:     1,
				SuccessNum: 10,
				RespNum:    10,
			},
		},
		{
			// L5：csv 格式不得 panic，且在表头中统一产出
			name: "CSV format stats",
			stats: &RequestStats{
				Format:     "csv",
				Url:        "http://example.com",
				RespSize:   1024,
				Duration:   time.Second * 5,
				MinReqTime: time.Millisecond * 100,
				MaxReqTime: time.Millisecond * 500,
				ErrNum:     1,
				SuccessNum: 10,
				RespNum:    10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 PrintStats 方法是否能正常执行而不 panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RequestStats.PrintStats() panicked: %v", r)
				}
			}()
			tt.stats.PrintStats()
		})
	}
}

func TestRequestStats_PrintStats_DivisionByZero(t *testing.T) {
	// 测试除零情况
	stats := &RequestStats{
		Format:     "raw",
		Url:        "http://example.com",
		RespSize:   1024,
		Duration:   0, // 零持续时间
		MinReqTime: time.Millisecond * 100,
		MaxReqTime: time.Millisecond * 500,
		ErrNum:     0,
		SuccessNum: 0,
		RespNum:    0,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RequestStats.PrintStats() panicked with division by zero: %v", r)
		}
	}()

	stats.PrintStats()
}

// captureStdout 捕获 fn 执行期间写向 stdout 的内容
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan string)
	go func() {
		out, _ := io.ReadAll(r)
		done <- string(out)
	}()
	fn()
	w.Close()
	os.Stdout = orig
	out := <-done // 先等读取 goroutine 拿到 EOF 收尾，再关 r，避免提前关闭读到空
	r.Close()
	return out
}

// TestPrintStats_AllFailedNoSentinelLeak 验证 L4：全部失败时
// 不得把 MaxInt64 哨兵值原样输出。
func TestPrintStats_AllFailedNoSentinelLeak(t *testing.T) {
	stats := &RequestStats{
		Format:     "raw",
		Url:        "http://example.com",
		MinReqTime: time.Duration(math.MaxInt64),
		ErrNum:     42,
	}

	out := captureStdout(t, stats.PrintStats)
	if strings.Contains(out, "2562047h") {
		t.Errorf("MaxInt64 sentinel leaked into output: %q", out)
	}
}

// TestPrintStats_JSONFriendlyFields 验证 Q2/L5：json 输出使用友好字段名、
// 毫秒浮点时长，且 stdout 拿到单行合法 JSON（旧实现不打印且时长是纳秒整数）。
func TestPrintStats_JSONFriendlyFields(t *testing.T) {
	stats := &RequestStats{
		Format:     "json",
		Url:        "http://example.com",
		RespSize:   1024,
		Duration:   time.Second * 5,
		MinReqTime: time.Millisecond * 100,
		MaxReqTime: time.Millisecond * 500,
		ErrNum:     1,
		SuccessNum: 10,
		RespNum:    2,
	}

	out := captureStdout(t, stats.PrintStats)

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON: %v, out=%q", err, out)
	}
	for _, key := range []string{"url", "threads", "success", "failed", "readKB", "requestsPerSec", "avgRequestMs", "maxRequestMs", "minRequestMs"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("json output missing field %q: %v", key, payload)
		}
	}
	if v, ok := payload["maxRequestMs"].(float64); !ok || v != 500 {
		t.Errorf("maxRequestMs = %v, want 500 (ms float, not ns integer)", payload["maxRequestMs"])
	}
}

// TestPrintStats_RPSIncludesFailures 验证吞吐语义：requestsPerSec 按总请求数
// （成功+失败）除以实际挂钟时长 Elapsed 计算，目标宕机（全失败）时不为 0。
func TestPrintStats_RPSIncludesFailures(t *testing.T) {
	stats := &RequestStats{
		Format:     "json",
		Url:        "http://example.com",
		Elapsed:    10 * time.Second,
		SuccessNum: 100,
		ErrNum:     50,
		RespSize:   10 * 1024,
		RespNum:    4,
	}

	out := captureStdout(t, stats.PrintStats)

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON: %v, out=%q", err, out)
	}
	if v, ok := payload["requestsPerSec"].(float64); !ok || v != 15.0 {
		t.Errorf("requestsPerSec = %v, want 15.0 ((100+50)/10s)", payload["requestsPerSec"])
	}
	if v, ok := payload["elapsedSeconds"].(float64); !ok || v != 10.0 {
		t.Errorf("elapsedSeconds = %v, want 10.0", payload["elapsedSeconds"])
	}

	// 全失败场景：吞吐仍反映发送速率
	allFailed := &RequestStats{Format: "json", Url: "http://example.com", Elapsed: 10 * time.Second, ErrNum: 200, RespNum: 2}
	out = captureStdout(t, allFailed.PrintStats)
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("all-failed json invalid: %v", err)
	}
	if v, ok := payload["requestsPerSec"].(float64); !ok || v != 20.0 {
		t.Errorf("all-failed requestsPerSec = %v, want 20.0", payload["requestsPerSec"])
	}
}

// TestPrintStats_CSVShape 验证 L5：csv 输出为「表头 + 一行数据」。
func TestPrintStats_CSVShape(t *testing.T) {
	stats := &RequestStats{
		Format:     "csv",
		Url:        "http://example.com,a", // 含逗号的 url 必须被引号包裹
		RespSize:   1024,
		Duration:   time.Second * 5,
		MinReqTime: time.Millisecond * 100,
		MaxReqTime: time.Millisecond * 500,
		ErrNum:     1,
		SuccessNum: 10,
		RespNum:    10,
	}

	out := captureStdout(t, stats.PrintStats)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("csv output must be header + 1 row, got %d lines: %q", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "url,threads,success,failed") {
		t.Errorf("unexpected csv header: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], `"http://example.com,a",`) {
		t.Errorf("url with comma must be quoted: %q", lines[1])
	}
}
