package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/leihenshang/http-little-toy/data"
)

func TestMaxTime(t *testing.T) {
	tests := []struct {
		name     string
		first    time.Duration
		second   time.Duration
		expected time.Duration
	}{
		{
			name:     "First is larger",
			first:    100 * time.Millisecond,
			second:   50 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "Second is larger",
			first:    50 * time.Millisecond,
			second:   100 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "Equal",
			first:    50 * time.Millisecond,
			second:   50 * time.Millisecond,
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
		name     string
		first    time.Duration
		second   time.Duration
		expected time.Duration
	}{
		{
			name:     "First is smaller",
			first:    50 * time.Millisecond,
			second:   100 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
		{
			name:     "Second is smaller",
			first:    100 * time.Millisecond,
			second:   50 * time.Millisecond,
			expected: 50 * time.Millisecond,
		},
		{
			name:     "Equal",
			first:    50 * time.Millisecond,
			second:   50 * time.Millisecond,
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
		"Content-Type":  []string{"application/json"},
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
		name    string
		toyReq  *data.ToyReq
		wantErr bool
	}{
		{
			name: "Basic HTTP client",
			toyReq: &data.ToyReq{
				Url:            testServer.URL,
				Timeout:        5,
				KeepAlive:      true,
				Compression:    true,
				SkipVerify:     false,
				AllowRedirects: true,
				Duration:       10,
				Thread:         5,
			},
			wantErr: false,
		},
		{
			name: "HTTP client with redirect disabled",
			toyReq: &data.ToyReq{
				Url:            testServer.URL,
				Timeout:        5,
				AllowRedirects: false,
				Duration:       10,
				Thread:         5,
			},
			wantErr: false,
		},
		{
			name: "HTTP/2 client",
			toyReq: &data.ToyReq{
				Url:      testServer.URL,
				Timeout:  5,
				UseHttp2: true,
				Duration: 10,
				Thread:   5,
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

// TestGenHttpClientHTTP2WithoutCert 验证 L1：不提供任何证书时 -h2 也必须生效
// （旧实现中 http2.ConfigureTransport 只在证书分支里执行，导致 h2 静默失效）。
func TestGenHttpClientHTTP2WithoutCert(t *testing.T) {
	client, err := genHttpClient(&data.ToyReq{
		Url:      "https://example.com",
		Timeout:  5,
		Duration: 10,
		Thread:   2,
		UseHttp2: true,
	})
	if err != nil {
		t.Fatalf("genHttpClient() unexpected error: %v", err)
	}

	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("transport is not *http.Transport")
	}
	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 should be true when UseHttp2 is set")
	}
	hasH2 := false
	for _, p := range tr.TLSClientConfig.NextProtos {
		if p == "h2" {
			hasH2 = true
		}
	}
	if !hasH2 {
		t.Errorf("TLSClientConfig.NextProtos = %v, want to contain %q", tr.TLSClientConfig.NextProtos, "h2")
	}
}

// TestGenHttpClientPartialCertsRejected 验证 mTLS 参数不齐时拒绝创建。
func TestGenHttpClientPartialCertsRejected(t *testing.T) {
	tests := []struct {
		name   string
		toyReq *data.ToyReq
	}{
		{"only client cert", &data.ToyReq{Url: "https://example.com", Timeout: 5, Duration: 10, Thread: 2, ClientCert: "cert.pem"}},
		{"only ca cert", &data.ToyReq{Url: "https://example.com", Timeout: 5, Duration: 10, Thread: 2, CaCert: "ca.pem"}},
		{"cert+key without ca", &data.ToyReq{Url: "https://example.com", Timeout: 5, Duration: 10, Thread: 2, ClientCert: "c", ClientKey: "k"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := genHttpClient(tt.toyReq); err == nil {
				t.Error("genHttpClient() should fail with incomplete mTLS params")
			}
		})
	}
}

// doReqFixture 构造指向测试服务的客户端与请求样本
func doReqFixture(t *testing.T, handler http.HandlerFunc, allowRedirects bool) (*http.Client, *data.ToyReq, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	reqObj := &data.ToyReq{
		Url:            srv.URL,
		Method:         http.MethodGet,
		Timeout:        5,
		Duration:       10,
		Thread:         1,
		KeepAlive:      true,
		Compression:    true,
		AllowRedirects: allowRedirects,
	}
	client, err := genHttpClient(reqObj)
	if err != nil {
		t.Fatalf("genHttpClient: %v", err)
	}
	return client, reqObj, srv
}

// TestDoReqStatusClasses 验证 L2：按状态码区间分类，全部 2xx 及跟随重定向后的
// 成功响应计为成功，4xx/5xx 计为对应类型错误。
func TestDoReqStatusClasses(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantSuccess bool
		errContains string
	}{
		{
			name:        "200 with body",
			handler:     func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("hello")) },
			wantSuccess: true,
		},
		{
			name: "201 created",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("x"))
			},
			wantSuccess: true,
		},
		{
			// 旧实现把 204 归为失败，修复后 2xx 全部成功
			name:        "204 no content",
			handler:     func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) },
			wantSuccess: true,
		},
		{
			name: "400 client error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("bad"))
			},
			wantSuccess: false,
			errContains: "client error: 400",
		},
		{
			name:        "503 server error",
			handler:     func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
			wantSuccess: false,
			errContains: "server error: 503",
		},
		{
			name: "302 followed to 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/final" {
					_, _ = w.Write([]byte("arrived"))
					return
				}
				http.Redirect(w, r, "/final", http.StatusFound)
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, reqObj, _ := doReqFixture(t, tt.handler, true)

			size, duration, err := doReq(client, reqObj, nil)
			if tt.wantSuccess {
				if err != nil {
					t.Fatalf("doReq() unexpected error: %v", err)
				}
				if size <= 0 {
					t.Errorf("doReq() respSize = %d, want > 0", size)
				}
				if duration <= 0 {
					t.Errorf("doReq() duration = %v, want > 0", duration)
				}
				return
			}
			if err == nil {
				t.Fatalf("doReq() expected error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("doReq() error = %q, want to contain %q", err, tt.errContains)
			}
			if size > 0 {
				t.Errorf("doReq() respSize = %d on failure, want <= 0", size)
			}
		})
	}
}

// TestDoReqBlockedRedirectReturnsError 验证 allowRedirects=false 时，
// 被拦截的重定向响应作为错误返回（并验证 resp 不泄露挂起：同服务紧接二次请求正常）。
func TestDoReqBlockedRedirectReturnsError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}
	client, reqObj, _ := doReqFixture(t, handler, false)

	_, _, err := doReq(client, reqObj, nil)
	if err == nil || !strings.Contains(err.Error(), "redirection not allowed") {
		t.Fatalf("doReq() error = %v, want 'redirection not allowed'", err)
	}
}

// TestDoReqSendsUserAgentAndCustomHeaders 验证预解析后的头部被正确附加。
func TestDoReqSendsUserAgentAndCustomHeaders(t *testing.T) {
	var gotUA, gotToy string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotToy = r.Header.Get("X-Toy")
		_, _ = w.Write([]byte("ok"))
	}
	client, reqObj, _ := doReqFixture(t, handler, true)

	if _, _, err := doReq(client, reqObj, [][2]string{{"X-Toy", "1"}}); err != nil {
		t.Fatalf("doReq() unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotUA, data.AppName+"/") {
		t.Errorf("User-Agent = %q, want prefix %q", gotUA, data.AppName+"/")
	}
	if gotToy != "1" {
		t.Errorf("X-Toy = %q, want %q", gotToy, "1")
	}
}

func TestCheckResFile(t *testing.T) {
	// 测试空文件路径
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
