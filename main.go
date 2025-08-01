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
	toyReq "github.com/leihenshang/http-little-toy/request"
)

var (
	respChan chan data.RequestStats
	helpTips = flag.Bool("h", false, "show help tips.")
	version  = flag.Bool("v", false, "show version.")
)

func initRequestSample() *data.RequestSample {
	requestSample := &data.RequestSample{}
	flag.Var(&requestSample.Params.Header, "header", "The http header.")
	flag.StringVar(&requestSample.Params.Url, "u", "", "The URL you want to test.")
	flag.StringVar(&requestSample.Params.Method, "m", http.MethodGet, "The http method.")
	flag.StringVar(&requestSample.Params.Body, "body", "", "The http body.")
	flag.IntVar(&requestSample.Params.Duration, "d", 10, "Duration of request.The unit is seconds.")
	flag.IntVar(&requestSample.Params.Thread, "t", 10, "Number of threads.")
	flag.BoolVar(&requestSample.Params.KeepAlive, "keepAlive", true, "Use keep-alive for http protocol.")
	flag.BoolVar(&requestSample.Params.Compression, "compression", true, "Use keep-alive for http protocol.")
	flag.IntVar(&requestSample.Params.Timeout, "timeout", 5, "the time out to wait response.the unit is seconds.")
	flag.BoolVar(&requestSample.Params.SkipVerify, "skipVerify", false, "TLS skipVerify.")
	flag.BoolVar(&requestSample.Params.AllowRedirects, "allowRedirects", true, "allowRedirects.")
	flag.BoolVar(&requestSample.Params.UseHttp2, "h2", false, "useHttp2.")
	flag.StringVar(&requestSample.Params.ClientCert, "clientCert", "", "clientCert.")
	flag.StringVar(&requestSample.Params.ClientKey, "clientKey", "", "clientKey.")
	flag.StringVar(&requestSample.Params.CaCert, "caCert", "", "caCert.")
	flag.Parse()

	return requestSample
}

func main() {
	reqSample := initRequestSample()
	reqSample.TipsAndHelp(*helpTips, *version)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	req, err := reqSample.GenReq()
	if err != nil {
		log.Fatal(err)
	}

	respChan = make(chan data.RequestStats, reqSample.Params.Thread)
	fmt.Printf("use %d coroutines,duration %d seconds.\n", reqSample.Params.Thread, reqSample.Params.Duration)
	fmt.Printf("%s %s header: %v \n", req.Method, req.Url, strings.Join(req.Header, "\n"))
	fmt.Println("---------------stats---------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(reqSample.Params.Duration)*time.Second)
	defer cancel()

	for i := 1; i <= reqSample.Params.Thread; i++ {
		go func() {
			httpCtx := context.Background()
			client, clientErr := toyReq.GetHttpClient(
				reqSample.Params.KeepAlive,
				reqSample.Params.Compression,
				time.Duration(reqSample.Params.Timeout),
				reqSample.Params.SkipVerify,
				reqSample.Params.AllowRedirects,
				reqSample.Params.ClientCert,
				reqSample.Params.ClientKey,
				reqSample.Params.CaCert,
				reqSample.Params.UseHttp2,
			)
			if clientErr != nil {
				log.Fatal(clientErr)
			}
			aggregate := data.RequestStats{MinReqTime: time.Hour}
		LOOP:
			for {
				size, d, _, err := toyReq.HandleReq(httpCtx, client, req)
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
	for allAggregate.RespNum < reqSample.Params.Thread {
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
