package data

import (
	"time"

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
	ReqUrl string `json:"reqUrl"`

	// header
	Headers xtype.StringSliceX `json:"headers"`

	// body
	Body string `json:"body"`

	// 日志文件
	LogFile string `json:"logFile"`

	// 持续时间
	Duration time.Duration `json:"duration"`

	// 线程数
	Thread int `json:"thread"`

	// 启用keep alive
	KeepAlive bool `json:"keepAlive"`

	// 启用压缩
	Compression bool `json:"compression"`

	// 请求文件
	RequestFile string `json:"requestFile"`

	// 创建请求文件模板
	GenerateSample bool `json:"generateSample"`

	// 等待响应超时时间
	TimeOut time.Duration `json:"timeOut"`

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

type RequestSample struct {
	ExecuteCount uint
	Params       Params
	Request      Request
}

//RequestStats 请求数据统计
type RequestStats struct {
	RespSize   int64
	Duration   time.Duration
	MinReqTime time.Duration
	MaxReqTime time.Duration
	ErrNum     int
	SuccessNum int
	RespNum    int
}

func (r Request) Valid() (err error) {
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
