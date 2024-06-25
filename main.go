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
	toyrequest "github.com/leihenshang/http-little-toy/request"
	"github.com/leihenshang/http-little-toy/sample"
)

var (
	respChan chan data.RequestStats
	toyLog   *mylog.ToyLog

	helpTips = flag.Bool("h", false, "show help tips.")
	version  = flag.Bool("v", false, "show version.")
)

func initRequestSample() *data.RequestSample {
	requestSample := new(data.RequestSample)
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
	flag.Parse()

	return requestSample
}

func main() {
	requestSample := initRequestSample()
	requestSample.TipsAndHelp(*helpTips, *version)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// generate a request sample
	if requestSample.Params.GenerateSample {
		err := sample.GenerateRequestFileV1("./request_sample.json", requestSample)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// check params object
	request, err := requestSample.ParseParams()
	if err != nil {
		log.Fatal(err)
	} else if err = request.Validate(); err != nil {
		log.Fatal(err)
	}

	logCtx, logCancel := context.WithCancel(context.Background())
	defer logCancel()
	toyLog = mylog.NewMyLog()
	if requestSample.Params.Log {
		if err = toyLog.Start(logCtx, data.LogDir); err != nil {
			log.Fatal(err)
		}
	}

	respChan = make(chan data.RequestStats, requestSample.Params.Thread)

	fmt.Printf("use %d coroutines,duration %d seconds.\n", requestSample.Params.Thread, requestSample.Params.Duration)
	fmt.Printf("url: %v method:%v header: %v \n", request.Url, request.Method, request.Header)
	fmt.Println("---------------stats---------------")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1e9*(requestSample.Params.Duration)))
	defer cancel()

	for i := 1; i <= requestSample.Params.Thread; i++ {
		go func() {
			httpCtx := context.Background()
			client, clientErr := toyrequest.GetHttpClient(
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

				if requestSample.Params.Log {
					toyLog.Wait.Add(1)
				}

				size, d, rawBody, err := toyrequest.HandleReq(httpCtx, client, request)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = timeUtil.MaxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = timeUtil.MinTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)

					if requestSample.Params.Log {
						// log write
						toyLog.Write(rawBody)
					}

				} else {
					log.Printf("request err:%+v\n", err)
					aggregate.ErrNum++
				}

				select {
				case <-ctx.Done():
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
		case <-signalChan:
			cancel()
		}
	}

	allAggregate.PrintStats()

	if requestSample.Params.Log {
		toyLog.Wait.Wait()
		d, _ := filepath.Abs(data.LogDir)
		log.Printf("log files are saves in:%+v \n", d)
	}

}
