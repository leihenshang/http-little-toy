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
	timeutil "tangzq/http-little-toy/common/utils/time-util"
	"tangzq/http-little-toy/model"
	reqObj "tangzq/http-little-toy/request"
	"tangzq/http-little-toy/sample"
	"time"
)

const Version = "0.0.1"
const Agent = "http-little-toy"

var respChan chan model.RequestStats

var duration = flag.Int("d", 0, "Duration of request.The unit is seconds.")
var thread = flag.Int("t", 0, "Number of threads.")
var keepAlive = flag.Bool("k", true, "Use keep-alive for http protocol.")
var compression = flag.Bool("compression", true, "Use keep-alive for http protocol.")
var reqUrl = flag.String("u", "", "The URL you want to test")
var requestFile = flag.String("f", "", "specify the request definition file.")
var generateSample = flag.Bool("g", false, "generate the request definition file template to the current directory.")
var version = flag.Bool("v", false, "show app version.")
var timeOut = flag.Uint("timeOut", 1000, "the time out to wait response")
var skipVerify = flag.Bool("skipVerify", false, "TLS skipVerify")
var allowRedirects = flag.Bool("allowRedirects", true, "allowRedirects")

var useHttp2 = flag.Bool("useHttp2", false, "useHttp2")

var clientCert = flag.String("clientCert", "", "clientCert")
var clientKey = flag.String("clientKey", "", "clientKey")
var caCert = flag.String("caCert", "", "caCert")

var helpTips = flag.Bool("h", false, "show help tips")

func printDefault() {
	fmt.Println("Usage: http-little-toy <options>")
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t", flag.Usage, "(Default "+flag.DefValue+")")
	})
}

func main() {
	// 设置一个信号通道，获取来自终端的终止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	flag.Parse()

	// 打印帮助
	if *helpTips {
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

	// 检查参数并使用 flag.Parse 解析命令行输入
	request := checkParams()

	// 请求校验
	validErr := request.Valid()
	if validErr != nil {
		log.Fatal(validErr)
	}

	fmt.Printf("use %d coroutines,duration %d seconds.", *thread, *duration)
	fmt.Printf("request: %+v\n", request)

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
				size, d, err := reqObj.HandleReq(httpCtx, client, request)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = timeutil.MaxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = timeutil.MinTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)
				} else {
					fmt.Println(err)
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
			allAggregate.MinReqTime = timeutil.MinTime(allAggregate.MinReqTime, r.MinReqTime)
			allAggregate.MaxReqTime = timeutil.MaxTime(allAggregate.MaxReqTime, r.MaxReqTime)
			respNum++
		case <-sigChan:
			cancel()
		}
	}

	averageThreadDuration := allAggregate.Duration / time.Duration(respNum)
	averageRequestTime := allAggregate.Duration / time.Duration(allAggregate.SuccessNum)
	perSecondTimes := float64(allAggregate.SuccessNum) / averageThreadDuration.Seconds()
	byteRate := float64(allAggregate.RespSize) / averageThreadDuration.Seconds()
	fmt.Printf("一共 %v 个请求,读取: %v KB \n", allAggregate.SuccessNum, allAggregate.RespSize/1024)
	fmt.Printf("requests/sec %.2f , Transfer/sec %.2f KB, average request time: %v \n", perSecondTimes, byteRate/1024, averageRequestTime)
	fmt.Printf("最慢的请求:%v \n", allAggregate.MaxReqTime)
	fmt.Printf("最快的请求:%v \n", allAggregate.MinReqTime)
	fmt.Printf("错误的请求数量：%v \n", allAggregate.ErrNum)
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

	if *requestFile != "" {
		fileBytes, err := ioutil.ReadFile(*requestFile)
		if err != nil {
			log.Fatal("an error occurred reading the file", err)
		}
		unmarshalErr := json.Unmarshal(fileBytes, &request)
		if unmarshalErr != nil {
			log.Fatal("unmarshal err: ", unmarshalErr)
		}
	}

	if *reqUrl != "" {
		request.Url = *reqUrl
		request.Method = http.MethodGet
	}

	return
}
