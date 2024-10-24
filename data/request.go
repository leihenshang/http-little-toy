package data

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/leihenshang/http-little-toy/common"
)

// Request 请求数据
type Request struct {
	Url    string   `json:"url"`
	Body   string   `json:"body"`
	Method string   `json:"method"`
	Header []string `json:"header"`
}

// Params 请求参数
type Params struct {
	// url
	Url string `json:"-"`

	// header
	Header MyStrSlice `json:"-"`

	// body
	Body string `json:"-"`

	// http 方法
	Method string `json:"-"`

	// 持续时间
	Duration int `json:"duration"`

	// 线程数
	Thread int `json:"thread"`

	// 启用keep alive
	KeepAlive bool `json:"keepAlive"`

	// 启用压缩
	Compression bool `json:"compression"`

	// 创建请求文件模板
	GenerateSample bool `json:"-"`

	// 等待响应超时时间
	TimeOut int `json:"timeOut"`

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

// RequestSample 请求模板对象
type RequestSample struct {
	// ExecuteCount uint
	Params  Params
	Request Request
}

// Validate 验证请求对象
func (r *Request) Validate() (err error) {
	// 检查 url 格式
	if urlErr := common.CheckUrl(r.Url); urlErr != nil {
		return urlErr
	}

	// 检查 method
	if methodErr := common.CheckHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}

	return
}

// ParseParams 解析参数
func (r *RequestSample) ParseParams() (req Request, err error) {
	if r.Params.Url == "" {
		err = errors.New("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL")
		return
	}

	req.Url = r.Params.Url
	req.Method = r.Params.Method
	req.Body = r.Params.Body
	req.Header = r.Params.Header

	return
}

// PrintDefault  打印默认操作
func (r *RequestSample) PrintDefault(appName string) {
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

func (r *RequestSample) TipsAndHelp(helpTips, version bool) {
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
