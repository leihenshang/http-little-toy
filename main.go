package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/leihenshang/http-little-toy/common"
	"github.com/leihenshang/http-little-toy/data"
	toyrequest "github.com/leihenshang/http-little-toy/request"
)

var (
	respChan chan data.RequestStats

	helpTips = flag.Bool("h", false, "show help tips.")
	version  = flag.Bool("v", false, "show version.")
)

func initRequestSample() *data.RequestSample {
	requestSample := new(data.RequestSample)
	flag.Var(&requestSample.Params.Header, "H", "The http header.")
	flag.StringVar(&requestSample.Params.Url, "u", "", "The URL you want to test.")
	flag.StringVar(&requestSample.Params.Method, "M", http.MethodGet, "The http method.")
	flag.StringVar(&requestSample.Params.Body, "body", "", "The http body.")
	flag.IntVar(&requestSample.Params.Duration, "d", 10, "Duration of request.The unit is seconds.")
	flag.IntVar(&requestSample.Params.Thread, "t", 10, "Number of threads.")
	flag.BoolVar(&requestSample.Params.KeepAlive, "keepAlive", true, "Use keep-alive for http protocol.")
	flag.BoolVar(&requestSample.Params.Compression, "compression", true, "Use keep-alive for http protocol.")
	flag.IntVar(&requestSample.Params.TimeOut, "timeOut", 5, "the time out to wait response.the unit is seconds.")
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

	request, err := requestSample.ParseParams()
	if err != nil {
		log.Fatal(err)
	} else if err = request.Validate(); err != nil {
		log.Fatal(err)
	}

	respChan = make(chan data.RequestStats, requestSample.Params.Thread)
	fmt.Printf("use %d coroutines,duration %d seconds.\n", requestSample.Params.Thread, requestSample.Params.Duration)
	fmt.Printf("%s %s header: %v \n", request.Method, request.Url, strings.Join(request.Header, "\n"))
	fmt.Println("---------------stats---------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestSample.Params.Duration)*time.Second)
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
				size, d, _, err := toyrequest.HandleReq(httpCtx, client, request)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = common.MaxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = common.MinTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)

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
			allAggregate.MinReqTime = common.MinTime(allAggregate.MinReqTime, r.MinReqTime)
			allAggregate.MaxReqTime = common.MaxTime(allAggregate.MaxReqTime, r.MaxReqTime)
			allAggregate.RespNum++
		case <-signalChan:
			cancel()
		}
	}

	allAggregate.PrintStats()
}
