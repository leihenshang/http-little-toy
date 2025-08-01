package data

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type ToyReq struct {
	// url
	Url string `json:"url"`

	// header
	Header MyStrSlice `json:"header"`

	// body
	Body string `json:"body"`

	// http Method
	Method string `json:"method"`

	// Duration 持续时间
	Duration int `json:"duration"`

	// Thread 线程数
	Thread int `json:"thread"`

	// KeepAlive 启用keep alive
	KeepAlive bool `json:"keepAlive"`

	// 启用压缩
	Compression bool `json:"compression"`

	// 创建请求文件模板
	GenerateSample bool `json:"-"`

	// 等待响应超时时间
	Timeout int `json:"timeout"`

	// 跳过TLS验证
	SkipVerify bool `json:"skipVerify"`

	// 允许重定向
	AllowRedirects bool `json:"allowRedirects"`

	// 使用http2
	UseHttp2 bool `json:"useHttp2"`

	// 客户端证书
	ClientCert string `json:"clientCert"`

	// 客户端秘钥
	ClientKey string `json:"clientKey"`

	// ca证书
	CaCert string `json:"caCert"`
}

func (r *ToyReq) GenReq() (err error) {
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
		http.MethodGet:    {},
		http.MethodHead:   {},
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodPatch:  {},
		http.MethodDelete: {},
		// http.MethodConnect,
		// http.MethodOptions,
		// http.MethodTrace,
	}
	if _, ok := httpMethodMap[method]; ok {
		return nil
	}
	return fmt.Errorf("%s is not in %s.", method, fmt.Sprintf("%v", httpMethodMap))
}
