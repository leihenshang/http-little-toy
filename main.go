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
	// 日志通道
	logChan chan []byte
	// 记录响应数据
	respChan chan data.RequestStats
	// 请求示例对象
	requestSample = new(data.RequestSample)
)

// 帮助
var helpTips = flag.Bool("h", false, "show help tips.")

// 版本打印
var version = flag.Bool("v", false, "show app version.")

func init() {
	flag.Var(&requestSample.Params.Headers, "H", "The http header.")
	flag.StringVar(&requestSample.Params.ReqUrl, "u", "", "The URL you want to test.")
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

//printDefault 打印默认操作
func printDefault() {
	fmt.Printf("Usage: %s <options>", AppName)
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
	// 解析所有标志
	flag.Parse()

	// 设置一个信号通道，获取来自终端的终止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// 打印帮助
	if *helpTips == true {
		printDefault()
		return
	}

	// 打印版本
	if *version == true {
		fmt.Printf("%s v%s \n", AppName, Version)
		return
	}

	// 创建请求模板
	if requestSample.Params.GenerateSample {
		err := sample.GenerateRequestFile("./request_sample.json")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	logCtx, logCancel := context.WithCancel(context.TODO())
	defer logCancel()
	if requestSample.Params.Log {
		logChan = make(chan []byte, requestSample.Params.Thread)
		logFile, logErr := myLog.CreateLog(LogDir)
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

	fmt.Printf("use %d coroutines,duration %d seconds.\n", requestSample.Params.Thread, requestSample.Params.Duration)
	fmt.Printf("url: %v method:%v header: %v \n", request.Url, request.Method, request.Header)
	fmt.Println("---------------stats---------------")

	// 使用该通道来存储请求的结果,并启用一个协程来读取该通道的结果
	respChan = make(chan data.RequestStats, requestSample.Params.Thread)

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
			// 日志协程退出
			logCancel()
		}
	}

	//打印结果
	printRes(allAggregate)

	if requestSample.Params.Log {
		// FIXME 不优雅的解决一下日志没写完的问题
		time.Sleep(2)
		logCancel()
		d, _ := filepath.Abs(LogDir)
		log.Printf("log in:%+v \n", d)
	}

}

func printRes(allAggregate data.RequestStats) {

	averageThreadDuration := func() time.Duration {
		if time.Duration(allAggregate.RespNum) <= 0 {
			return 0
		}
		return allAggregate.Duration / time.Duration(allAggregate.RespNum)

	}()
	averageRequestTime := func() time.Duration {
		if time.Duration(allAggregate.SuccessNum) <= 0 {
			return 0
		}

		return allAggregate.Duration / time.Duration(allAggregate.SuccessNum)
	}()
	perSecondTimes := float64(allAggregate.SuccessNum) / averageThreadDuration.Seconds()
	byteRate := float64(allAggregate.RespSize) / averageThreadDuration.Seconds()

	fmt.Printf("number of success: %v ,number of failed: %v,read: %v KB \n", allAggregate.SuccessNum, allAggregate.ErrNum, allAggregate.RespSize/1024)
	fmt.Printf("requests/sec %.2f , transfer/sec %.2f KB, average request time: %v \n", perSecondTimes, byteRate/1024, averageRequestTime)
	fmt.Printf("the slowest request:%v \n", allAggregate.MaxReqTime)
	fmt.Printf("the fastest request:%v \n", allAggregate.MinReqTime)

}

func checkParams() (request data.Request) {
	if requestSample.Params.RequestFile == "" && requestSample.Params.ReqUrl == "" {
		log.Fatal("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL.")
	}

	if requestSample.Params.RequestFile != "" && requestSample.Params.ReqUrl != "" {
		log.Fatal("the \"-u\" or \"-f\" parameter can not exist the same time.")
	}

	// 默认请求文件优先级最高
	if requestSample.Params.RequestFile != "" {
		fileBytes, err := ioutil.ReadFile(requestSample.Params.RequestFile)
		if err != nil {
			log.Fatal("an error occurred reading the file", err)
		}
		unmarshalErr := json.Unmarshal(fileBytes, &request)
		if unmarshalErr != nil {
			log.Fatal("unmarshal err: ", unmarshalErr)
		}

	} else {
		request.Url = requestSample.Params.ReqUrl
		request.Method = http.MethodGet
		request.Body = requestSample.Params.Body
		request.Header = requestSample.Params.Headers
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
