package data

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
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

	// 超时时间 Timeout
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

func (r *ToyReq) Check() (err error) {
	if r.Url == "" {
		err = errors.New("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL")
		return
	}

	if methodErr := checkHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}
	return nil
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

func checkHttpMethod(method string) error {
	var httpMethodMap = map[string]struct{}{
		http.MethodGet:     {},
		http.MethodHead:    {},
		http.MethodPost:    {},
		http.MethodPut:     {},
		http.MethodPatch:   {},
		http.MethodDelete:  {},
		http.MethodConnect: {},
		http.MethodOptions: {},
		http.MethodTrace:   {},
	}
	if _, ok := httpMethodMap[method]; ok {
		return nil
	}
	return fmt.Errorf("%s is not in %s.", method, fmt.Sprintf("%v", httpMethodMap))
}
