package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/leihenshang/http-little-toy/common/mylog"
	timeUtil "github.com/leihenshang/http-little-toy/common/utils/time-util"
	"github.com/leihenshang/http-little-toy/data"
	reqObj "github.com/leihenshang/http-little-toy/request"
	"github.com/leihenshang/http-little-toy/sample"
)

// ----基本设置----
const (
	// 版本
	Version = "0.0.3"

	// 请求代理名称
	AppName = "http-little-toy"

	// 日志目录
	LogDir = "log"
)

// ----数据交换----
var (
	// 记录响应数据
	respChan chan data.RequestStats
	// 日志
	mLog *mylog.MyLog
	// 请求示例对象
	requestSample = new(data.RequestSample)
)

// 帮助
var helpTips = flag.Bool("h", false, "show help tips.")

// 版本打印
var version = flag.Bool("v", false, "show app version.")

func init() {
	flag.Var(&requestSample.Params.Header, "H", "The http header.")
	flag.StringVar(&requestSample.Params.Url, "u", "", "The URL you want to test.")
	flag.StringVar(&requestSample.Params.Method, "M", http.MethodGet, "The http method.")
	flag.StringVar(&requestSample.Params.Body, "body", "", "The http body.")
	flag.BoolVar(&requestSample.Params.Log, "log", false, "Log the request response to file. default: './log'")
	flag.IntVar(&requestSample.Params.Duration, "d", 10, "Duration of request.The unit is seconds.")
	flag.IntVar(&requestSample.Params.Thread, "t", 10, "Number of threads.")
	flag.BoolVar(&requestSample.Params.KeepAlive, "keepAlive", true, "Use keep-alive for http protocol.")
	flag.BoolVar(&requestSample.Params.Compression, "compression", true, "Use keep-alive for http protocol.")
	flag.StringVar(&requestSample.Params.RequestFile, "f", "", "specify the request definition file.")
	flag.BoolVar(&requestSample.Params.GenerateSample, "gen", false, "generate the request definition file template to the current directory.")
	flag.IntVar(&requestSample.Params.TimeOut, "timeOut", 1000, "the time out to wait response.")
	flag.BoolVar(&requestSample.Params.SkipVerify, "skipVerify", false, "TLS skipVerify.")
	flag.BoolVar(&requestSample.Params.AllowRedirects, "allowRedirects", true, "allowRedirects.")
	flag.BoolVar(&requestSample.Params.UseHttp2, "useHttp2", false, "useHttp2.")
	flag.StringVar(&requestSample.Params.ClientCert, "clientCert", "", "clientCert.")
	flag.StringVar(&requestSample.Params.ClientKey, "clientKey", "", "clientKey.")
	flag.StringVar(&requestSample.Params.CaCert, "caCert", "", "caCert.")
}

func main() {
	// 解析所有标志
	flag.Parse()

	// 设置一个信号通道，获取来自终端的终止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// 打印帮助
	if *helpTips {
		requestSample.PrintDefault(AppName)
		return
	}

	// 打印版本
	if *version {
		fmt.Printf("%s v%s \n", AppName, Version)
		return
	}

	// 创建请求模板
	if requestSample.Params.GenerateSample {
		err := sample.GenerateRequestFileV1("./request_sample.json", requestSample)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// 检查参数对象
	request, parseErr := requestSample.ParseParams()
	if parseErr != nil {
		log.Fatal(parseErr)
	}
	// 请求对象校验
	validErr := request.Valid()
	if validErr != nil {
		log.Fatal(validErr)
	}

	logCtx, logCancel := context.WithCancel(context.TODO())
	defer logCancel()
	if requestSample.Params.Log {
		mLog = mylog.NewMyLog()
		logErr := mLog.LogStart(logCtx, LogDir)
		if logErr != nil {
			log.Fatal(logErr)
		}
	}

	// 初始化通道
	respChan = make(chan data.RequestStats, requestSample.Params.Thread)

	fmt.Printf("use %d coroutines,duration %d seconds.\n", requestSample.Params.Thread, requestSample.Params.Duration)
	fmt.Printf("url: %v method:%v header: %v \n", request.Url, request.Method, request.Header)
	fmt.Println("---------------stats---------------")

	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1e9*(requestSample.Params.Duration)))
	defer cancel()

	for i := 1; i <= requestSample.Params.Thread; i++ {
		go func() {
			httpCtx := context.TODO()
			client, clientErr := reqObj.GetHttpClient(
				requestSample.Params.KeepAlive,
				requestSample.Params.Compression,
				time.Duration(requestSample.Params.TimeOut),
				requestSample.Params.SkipVerify,
				requestSample.Params.AllowRedirects,
				requestSample.Params.ClientCert,
				requestSample.Params.ClientKey,
				requestSample.Params.CaCert,
				requestSample.Params.UseHttp2,
			)
			if clientErr != nil {
				log.Fatal(clientErr)
			}
			aggregate := data.RequestStats{MinReqTime: time.Hour}
		LOOP:
			for {
				size, d, rawBody, err := reqObj.HandleReq(httpCtx, client, request)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = timeUtil.MaxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = timeUtil.MinTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)

					if requestSample.Params.Log {
						//写入日志通道
						mLog.WriteLog(rawBody)
					}

				} else {
					log.Printf("request err:%+v\n", err)
					aggregate.ErrNum++
				}

				select {
				case <-ctx.Done():
					// 结束执行
					break LOOP
				default:
				}
			}
			respChan <- aggregate
		}()
	}

	allAggregate := data.RequestStats{MinReqTime: time.Hour}
	for allAggregate.RespNum < requestSample.Params.Thread {
		select {
		case r := <-respChan:
			allAggregate.ErrNum += r.ErrNum
			allAggregate.SuccessNum += r.SuccessNum
			allAggregate.RespSize += r.RespSize
			allAggregate.Duration += r.Duration
			allAggregate.MinReqTime = timeUtil.MinTime(allAggregate.MinReqTime, r.MinReqTime)
			allAggregate.MaxReqTime = timeUtil.MaxTime(allAggregate.MaxReqTime, r.MaxReqTime)
			allAggregate.RespNum++
		case <-sigChan:
			cancel()
		}
	}

	//打印结果
	allAggregate.PrintStats()

	if requestSample.Params.Log {
		// FIXME 不优雅的解决一下日志没写完的问题
		time.Sleep(time.Second * 2)
		d, _ := filepath.Abs(LogDir)
		log.Printf("log in:%+v \n", d)
	}

}
