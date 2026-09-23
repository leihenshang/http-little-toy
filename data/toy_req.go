package data

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ToyReq struct {
	// 目标地址 Url target url
	Url string `json:"url"`

	// HTTP头 Header
	Header MyStrSlice `json:"header"`

	// 请求体 Body request body
	Body string `json:"body"`

	// HTTP方法 Method http Method
	Method string `json:"method"`

	// 请求持续时间 Duration time
	Duration int `json:"duration"`

	// 线程数（模拟的用户数量） Thread is the number of threads
	Thread int `json:"thread"`

	// 是否启用保持连接  keep-alive
	KeepAlive bool `json:"keepAlive"`

	// 是否启用压缩 Compression
	Compression bool `json:"compression"`

	// http请求的总超时时间 Timeout
	Timeout int `json:"timeout"`

	// 是否跳过TLS验证 SkipVerify is whether to skip TLS verification
	SkipVerify bool `json:"skipVerify"`

	// 是否允许重定向 AllowRedirects
	AllowRedirects bool `json:"allowRedirects"`

	// 是否启用HTTP/2 UseHttp2
	UseHttp2 bool `json:"useHttp2"`

	// 客户端证书 ClientCert
	ClientCert string `json:"clientCert"`

	// 客户端密钥 ClientKey
	ClientKey string `json:"clientKey"`

	// CA证书 CaCert
	CaCert string `json:"caCert"`
}

const (
	// MaxThread 并发上限，防止误设超大值自我 DoS（L3）
	// MaxThread bounds concurrency to prevent accidental resource exhaustion.
	MaxThread = 10000
	// MaxDuration 单次测试时长上限（秒）/ max test duration in seconds
	MaxDuration = 24 * 3600
)

func (r *ToyReq) Validate() (err error) {
	if r.Url == "" {
		err = errors.New("Use the \"-u\" parameter to set the URL. eg: http-little-toy -u https://example.com")
		return
	}

	if methodErr := checkHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}

	// L3：数值参数边界校验，消除 make(chan, 负数) panic 与无界并发
	if r.Thread <= 0 || r.Thread > MaxThread {
		return fmt.Errorf("-t must be in [1, %d], got %d", MaxThread, r.Thread)
	}
	if r.Duration <= 0 || r.Duration > MaxDuration {
		return fmt.Errorf("-d must be in [1, %d] seconds, got %d", MaxDuration, r.Duration)
	}
	if r.Timeout <= 0 {
		return fmt.Errorf("-timeout must be positive, got %d", r.Timeout)
	}

	// S2：自定义头部启动期校验，非法/危险项直接拒绝
	if _, hdrErr := r.ParseHeaders(); hdrErr != nil {
		return hdrErr
	}
	return nil
}

// ParseHeaders 把 ["Key: value", ...] 解析为键值对并做安全校验（S2）：
// 头部名必须是合法 token，名/值均不得包含 CR/LF，防止头注入。
// ParseHeaders parses and sanitizes custom headers; CR/LF and invalid
// field names are rejected fail-fast.
func (r *ToyReq) ParseHeaders() ([][2]string, error) {
	parsed := make([][2]string, 0, len(r.Header))
	for _, v := range r.Header {
		kv := strings.SplitN(v, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid header %q: missing ':' separator", v)
		}
		key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		if !isHeaderFieldName(key) {
			return nil, fmt.Errorf("invalid header %q: bad header name %q (empty, space or CR/LF not allowed)", v, key)
		}
		if strings.ContainsAny(value, "\r\n") {
			return nil, fmt.Errorf("invalid header %q: CR/LF in header value is not allowed", v)
		}
		parsed = append(parsed, [2]string{key, value})
	}
	return parsed, nil
}

// isHeaderFieldName 按 RFC 7230 tchar 规则校验头部名 / validates header name
func isHeaderFieldName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isToken := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			strings.IndexByte("!#$%&'*+-.^_`|~", c) >= 0
		if !isToken {
			return false
		}
	}
	return true
}

func (r *ToyReq) PrintDefault(appName string) {
	fmt.Printf("Usage: %s <options>", appName)
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t\n\t\t", flag.Usage, "--default="+func() string {
			if flag.DefValue == "" {
				return "\"\""
			}

			return flag.DefValue
		}()+".")
	})
}

func (r *ToyReq) TipsAndHelp(helpTips, version bool) {
	if helpTips {
		r.PrintDefault(AppName)
		os.Exit(0)
	}

	if version {
		fmt.Printf("%s v%s \n", AppName, Version)
		os.Exit(0)
	}
}

type MyStrSlice []string

func (s *MyStrSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *MyStrSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// allowedMethods 有序列表：校验同时用于错误提示（Q1：替代无序 map 的打印）
var allowedMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions,
}

func checkHttpMethod(method string) error {
	for _, m := range allowedMethods {
		if method == m {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid HTTP method, allowed: %s", method, strings.Join(allowedMethods, ", "))
}
