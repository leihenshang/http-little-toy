package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leihenshang/http-little-toy/data"
)

func TestMaxTime(t *testing.T) {
	tests := []struct {
		name  string
		first time.Duration
		second time.Duration
		expected time.Duration
	}{
		{
			name:  "First is larger",
			first: 100 * time.Millisecond,
			second: 50 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
		{
			name:  "Second is larger",
			first: 50 * time.Millisecond,
			second: 100 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
		{
			name:  "Equal",
			first: 50 * time.Millisecond,
			second: 50 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxTime(tt.first, tt.second)
			if result != tt.expected {
				t.Errorf("maxTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMinTime(t *testing.T) {
	tests := []struct {
		name  string
		first time.Duration
		second time.Duration
		expected time.Duration
	}{
		{
			name:  "First is smaller",
			first: 50 * time.Millisecond,
			second: 100 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
		{
			name:  "Second is smaller",
			first: 100 * time.Millisecond,
			second: 50 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
		{
			name:  "Equal",
			first: 50 * time.Millisecond,
			second: 50 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minTime(tt.first, tt.second)
			if result != tt.expected {
				t.Errorf("minTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateHttpHeadersSize(t *testing.T) {
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"Authorization": []string{"Bearer token123"},
	}
	result := calculateHttpHeadersSize(headers)
	// 计算预期大小
	// Content-Type: application/json -> 24 + 16 = 40
	// Authorization: Bearer token123 -> 13 + 14 = 27  
	// \r\n -> 2
	expected := int64(66)
	if result != expected {
		t.Errorf("calculateHttpHeadersSize() = %v, want %v", result, expected)
	}
}

func TestGenHttpClient(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer testServer.Close()

	tests := []struct {
		name     string
		toyReq   *data.ToyReq
		wantErr  bool
	}{
		{
			name: "Basic HTTP client",
			toyReq: &data.ToyReq{
				Url: testServer.URL,
				Timeout: 5,
				KeepAlive: true,
				Compression: true,
				SkipVerify: false,
				AllowRedirects: true,
				Duration: 10,
				Thread: 5,
			},
			wantErr: false,
		},
		{
			name: "HTTP client with redirect disabled",
			toyReq: &data.ToyReq{
				Url: testServer.URL,
				Timeout: 5,
				AllowRedirects: false,
				Duration: 10,
				Thread: 5,
			},
			wantErr: false,
		},
		{
			name: "HTTP/2 client",
			toyReq: &data.ToyReq{
				Url: testServer.URL,
				Timeout: 5,
				UseHttp2: true,
				Duration: 10,
				Thread: 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := genHttpClient(tt.toyReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("genHttpClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if client == nil {
					t.Error("genHttpClient() returned nil client")
					return
				}
				// 测试客户端是否能正常发送请求
				req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
				if err != nil {
					t.Errorf("Failed to create request: %v", err)
					return
				}
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("Client request failed: %v", err)
					return
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			}
		})
	}
}

func TestCheckResFile(t *testing.T) {
	// 测试空文件路径
	resFile = new(string)
	*resFile = ""
	file, err := checkResFile()
	if err != nil {
		t.Errorf("checkResFile() error = %v, want nil", err)
	}
	if file != nil {
		t.Error("checkResFile() should return nil for empty file path")
	}
}

func TestPrintLByFormat(t *testing.T) {
	// 测试不同格式的输出
	formats := []string{"raw", "json", "csv"}
	for _, format := range formats {
		t.Run("Format: "+format, func(t *testing.T) {
			// 测试 printLByFormat 方法是否能正常执行而不 panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printLByFormat() panicked: %v", r)
				}
			}()
			printLByFormat(format, "test message")
		})
	}
}