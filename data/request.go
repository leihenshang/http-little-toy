package data

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	httputil "github.com/leihenshang/http-little-toy/common/utils/http-util"
	"github.com/leihenshang/http-little-toy/common/xtype"
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
	Header xtype.StringSliceX `json:"-"`

	// body
	Body string `json:"-"`

	// 日志文件
	Log bool `json:"log"`

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

	// 请求文件
	RequestFile string `json:"-"`

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

//RequestSample 请求模板对象
type RequestSample struct {
	// ExecuteCount uint
	Params  Params
	Request Request
}

//Valid 验证请求对象
func (r *Request) Valid() (err error) {
	// 检查 url 格式
	if urlErr := httputil.CheckUrlAddr(r.Url); urlErr != nil {
		return urlErr
	}

	// 检查 method
	if methodErr := httputil.CheckHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}

	return
}

//ParseParams 解析参数
func (r *RequestSample) ParseParams() (reqObj Request, err error) {

	if r.Params.RequestFile == "" && r.Params.Url == "" {
		err = errors.New("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL")
		return
	}

	if r.Params.RequestFile != "" && r.Params.Url != "" {
		err = errors.New("the \"-u\" or \"-f\" parameter can not exist the same time")
		return
	}

	// 默认请求文件优先级最高
	if r.Params.RequestFile != "" {
		log.Printf("ParseParams: use request file: %s \n", r.Params.RequestFile)
		fileBytes, readErr := os.ReadFile(r.Params.RequestFile)
		if err != nil {
			err = errors.New("an error occurred reading the 'request_sample.json' file.err:" + readErr.Error())
			return
		}
		unmarshalErr := json.Unmarshal(fileBytes, &r)
		if unmarshalErr != nil {
			err = errors.New("unmarshal err: " + unmarshalErr.Error())
			return
		}
		// 请求文件参数

		reqObj.Url = r.Request.Url
		reqObj.Method = r.Request.Method
		reqObj.Body = r.Request.Body
		reqObj.Header = r.Request.Header

	} else {
		// 命令行参数

		reqObj.Url = r.Params.Url
		reqObj.Method = r.Params.Method
		reqObj.Body = r.Params.Body
		reqObj.Header = r.Params.Header
	}

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
