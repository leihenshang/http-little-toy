package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/leihenshang/http-little-toy/data"
	"golang.org/x/net/http2"
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
	fmt.Println("---------------stats---------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(reqSample.Params.Duration)*time.Second)
	defer cancel()

	for i := 1; i <= reqSample.Params.Thread; i++ {
		go func() {
			client, clientErr := GenHttpClient(
				reqSample.Params.KeepAlive,
				reqSample.Params.Compression,
				time.Duration(reqSample.Params.Timeout)*time.Second,
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
				size, d, _, err := HandleReq(client, req)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = maxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = minTime(aggregate.MinReqTime, d)
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
			allAggregate.MinReqTime = minTime(allAggregate.MinReqTime, r.MinReqTime)
			allAggregate.MaxReqTime = maxTime(allAggregate.MaxReqTime, r.MaxReqTime)
			allAggregate.RespNum++
		case <-signalChan:
			cancel()
		}
	}

	allAggregate.PrintStats()
}

func maxTime(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}
	return second
}

func minTime(first, second time.Duration) time.Duration {
	if first < second {
		return first
	}
	return second
}

func GenHttpClient(
	keepAlive bool,
	compression bool,
	timeout time.Duration,
	skipVerify bool,
	allowRedirects bool,
	clientCert string,
	clientKey string,
	caCert string,
	useHttp2 bool,
) (client *http.Client, err error) {
	client = &http.Client{}

	disableKeepAlive := !keepAlive
	disableCompression := !compression

	client.Transport = &http.Transport{
		ResponseHeaderTimeout: timeout,
		DisableCompression:    disableCompression,
		DisableKeepAlives:     disableKeepAlive,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipVerify},
	}

	if !allowRedirects {
		//returning an error when trying to redirect. This prevents the redirection from happening.
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirection not allowed")
		}
	}

	if clientCert == "" && clientKey == "" && caCert == "" {
		return client, nil
	}

	if clientCert == "" {
		return nil, fmt.Errorf("client certificate can't be empty")
	}
	if clientKey == "" {
		return nil, fmt.Errorf("client key can't be empty")
	}
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("unable to load cert tried to load %v and %v but got %v", clientCert, clientKey, err)
	}

	// load our CA certificate
	clientCACert, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("unable to open cert %v", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: skipVerify,
	}

	t := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if useHttp2 {
		if err = http2.ConfigureTransport(t); err != nil {
			return nil, err
		}
	}

	client.Transport = t
	return client, nil
}

func HandleReq(client *http.Client, reqObj data.Request) (respSize int, duration time.Duration, bodyBytes []byte, err error) {
	respSize = -1
	duration = -1

	req, err := http.NewRequest(reqObj.Method, reqObj.Url, strings.NewReader(reqObj.Body))
	if err != nil {
		fmt.Printf("new request failed, err:%v\n", err)
		return
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", data.AppName, data.Version))

	for _, v := range reqObj.Header {
		if temp := strings.SplitN(v, ":", 2); len(temp) == 2 {
			req.Header.Add(temp[0], temp[1])
		} else {
			fmt.Printf("split header error,value:%+v,split len:%v", v, len(temp))
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	duration = time.Since(start)

	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("an error occurred doing request(io readAll):", err)
	}

	headerSize := 0
	if len(resp.Header) > 0 {
		headerSize = int(calculateHttpHeadersSize(resp.Header))
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		respSize = len(bodyBytes) + headerSize
	case http.StatusMovedPermanently, http.StatusTemporaryRedirect:
		respSize = int(resp.ContentLength) + headerSize
	default:
		err = errors.New(fmt.Sprint("http-code:", resp.StatusCode, ",header: ", resp.Header, ",content: ", string(bodyBytes)))
	}

	return
}

func calculateHttpHeadersSize(headers http.Header) (result int64) {
	for k, v := range headers {
		result += int64(len(k) + len(": \r\n"))
		for _, s := range v {
			result += int64(len(s))
		}
	}
	result += int64(len("\r\n"))
	return result
}
