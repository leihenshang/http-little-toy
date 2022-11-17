package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	myLog "github.com/leihenshang/http-little-toy/common/mylog"
	timeUtil "github.com/leihenshang/http-little-toy/common/utils/time-util"
	"github.com/leihenshang/http-little-toy/common/xtype"
	"github.com/leihenshang/http-little-toy/model"
	reqObj "github.com/leihenshang/http-little-toy/request"
	"github.com/leihenshang/http-little-toy/sample"
)

// 版本
const Version = "0.0.2"

// 请求代理名称
const Agent = "http-little-toy"

// 日志目录
const LogDir = "log"

var logChan chan []byte

var respChan chan model.RequestStats

// 帮助
var helpTips = flag.Bool("h", false, "show help tips.")

// 版本打印
var version = flag.Bool("v", false, "show app version.")

// url
var reqUrl = flag.String("u", "", "The URL you want to test.")

// header
var headers xtype.StringSliceX

// body
var body = flag.String("body", "", "The http body.")

// 日志文件
var logFile = flag.Bool("log", false, "record request log to file. default: './log'")

// 持续时间
var duration = flag.Int("d", 0, "Duration of request.The unit is seconds.")

// 线程数
var thread = flag.Int("t", 0, "Number of threads.")

// 启用keep alive
var keepAlive = flag.Bool("keepAlive", true, "Use keep-alive for http protocol.")

// 启用压缩
var compression = flag.Bool("compression", true, "Use keep-alive for http protocol.")

// 请求文件
var requestFile = flag.String("f", "", "specify the request definition file.")

// 创建请求文件模板
var generateSample = flag.Bool("gen", false, "generate the request definition file template to the current directory.")

// 等待响应超时时间
var timeOut = flag.Uint("timeOut", 1000, "the time out to wait response.")

// 跳过TLS验证
var skipVerify = flag.Bool("skipVerify", false, "TLS skipVerify.")

// 允许重定向
var allowRedirects = flag.Bool("allowRedirects", true, "allowRedirects.")

// 使用http2
var useHttp2 = flag.Bool("useHttp2", false, "useHttp2.")

// 客户端证书
var clientCert = flag.String("clientCert", "", "clientCert.")

// 客户端秘钥
var clientKey = flag.String("clientKey", "", "clientKey.")

// ca证书
var caCert = flag.String("caCert", "", "caCert.")

func printDefault() {
	fmt.Println("Usage: httpToy <options>")
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

func main() {

	flag.Var(&headers, "H", "The http header.")

	// 设置一个信号通道，获取来自终端的终止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	flag.Parse()

	// 打印帮助
	if *helpTips == true {
		printDefault()
		return
	}

	// 打印版本
	if *version == true {
		fmt.Println("http-little-toy v" + Version)
		return
	}

	// 创建请求模板
	if *generateSample {
		err := sample.GenerateRequestFile("./request_sample.json")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	logCtx, logCancel := context.WithCancel(context.TODO())
	defer logCancel()
	if *logFile {
		logChan = make(chan []byte, *thread)
		logFile, logErr := myLog.GenLog(LogDir)
		if logErr != nil {
			log.Fatalf("an error occurred while get log file.err:%+v\n", logErr)
		}

		// 启动一个协程来处理日志写入
		go func(c context.Context) {
		LOOP:
			for {
				select {
				case l := <-logChan:
					logData := []byte(time.Now().Format("2006-01-02 15:04:05 "))
					logData = append(logData, l...)
					logData = append(logData, []byte("\n")...)
					_, lErr := logFile.Write(logData)
					if lErr != nil {
						log.Printf("write log err:%+v\n", lErr)
					}
				case <-c.Done():
					break LOOP
				}

			}
			// 关闭日志文件
			logFile.Close()
		}(logCtx)

	}

	// 检查参数并使用 flag.Parse 解析命令行输入
	request := checkParams()

	// 请求校验
	validErr := request.Valid()
	if validErr != nil {
		log.Fatal(validErr)
	}

	fmt.Printf("use %d coroutines,duration %d seconds.\n", *thread, *duration)
	fmt.Printf("url: %v method:%v header: %v \n", request.Url, request.Method, request.Header)
	fmt.Println("---------------stats---------------")

	// 使用该通道来存储请求的结果,并启用一个协程来读取该通道的结果
	respChan = make(chan model.RequestStats, *thread)

	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1e9*(*duration)))
	defer cancel()

	for i := 1; i <= *thread; i++ {
		go func() {
			httpCtx := context.TODO()
			client, clientErr := reqObj.GetHttpClient(
				*keepAlive,
				*compression,
				time.Duration(*timeOut),
				*skipVerify,
				*allowRedirects,
				*clientCert,
				*clientKey,
				*caCert,
				*useHttp2,
			)
			if clientErr != nil {
				log.Fatal(clientErr)
			}
			aggregate := model.RequestStats{MinReqTime: time.Hour}
		LOOP:
			for {
				size, d, rawBody, err := reqObj.HandleReq(httpCtx, client, request)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = timeUtil.MaxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = timeUtil.MinTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)

					if *logFile {
						//写入日志通道
						logChan <- rawBody
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

	respNum := 0
	allAggregate := model.RequestStats{MinReqTime: time.Hour}
	for respNum < *thread {
		select {
		case r := <-respChan:
			allAggregate.ErrNum += r.ErrNum
			allAggregate.SuccessNum += r.SuccessNum
			allAggregate.RespSize += r.RespSize
			allAggregate.Duration += r.Duration
			allAggregate.MinReqTime = timeUtil.MinTime(allAggregate.MinReqTime, r.MinReqTime)
			allAggregate.MaxReqTime = timeUtil.MaxTime(allAggregate.MaxReqTime, r.MaxReqTime)
			respNum++
		case <-sigChan:
			cancel()
			// 日志协程退出
			logCancel()
		}
	}

	averageThreadDuration := allAggregate.Duration / time.Duration(respNum)
	averageRequestTime := allAggregate.Duration / time.Duration(allAggregate.SuccessNum)
	perSecondTimes := float64(allAggregate.SuccessNum) / averageThreadDuration.Seconds()
	byteRate := float64(allAggregate.RespSize) / averageThreadDuration.Seconds()
	fmt.Printf("number of success: %v ,number of failed: %v,read: %v KB \n", allAggregate.SuccessNum, allAggregate.ErrNum, allAggregate.RespSize/1024)
	fmt.Printf("requests/sec %.2f , transfer/sec %.2f KB, average request time: %v \n", perSecondTimes, byteRate/1024, averageRequestTime)
	fmt.Printf("the slowest request:%v \n", allAggregate.MaxReqTime)
	fmt.Printf("the fastest request:%v \n", allAggregate.MinReqTime)

	// FIXME 不优雅的解决一下日志没写完的问题
	time.Sleep(2)
	logCancel()
	d, _ := filepath.Abs(LogDir)
	log.Printf("log in:%+v \n", d)
}

func checkParams() (request model.Request) {
	if *duration == 0 || *thread == 0 {
		log.Fatal("params is error.Use \"-d\" and \"-t\" parameter add the set.")
	}
	if *requestFile == "" && *reqUrl == "" {
		log.Fatal("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL.")
	}

	if *requestFile != "" && *reqUrl != "" {
		log.Fatal("the \"-u\" or \"-f\" parameter can not exist the same time.")
	}

	// 默认请求文件优先级最高
	if *requestFile != "" {
		fileBytes, err := ioutil.ReadFile(*requestFile)
		if err != nil {
			log.Fatal("an error occurred reading the file", err)
		}
		unmarshalErr := json.Unmarshal(fileBytes, &request)
		if unmarshalErr != nil {
			log.Fatal("unmarshal err: ", unmarshalErr)
		}

	} else {
		request.Url = *reqUrl
		request.Method = http.MethodGet
		request.Body = *body
		request.Header = headers
	}

	// fmt.Printf("%+v \n", request)
	// if len(request.Header) > 0 {
	// 	for _, v := range request.Header {
	// 		temp := strings.SplitN(v, ":", 2)
	// 		if len(temp) == 2 {
	// 			fmt.Printf("original:%+v, key:%+v,value:%+v \n", temp, temp[0], temp[1])
	// 		} else {
	// 			fmt.Printf("split header err,value:%+v,split len:%+v", v, len(temp))
	// 		}
	// 	}
	// }
	// os.Exit(0)
	return
}
