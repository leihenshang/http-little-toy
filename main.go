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
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/leihenshang/http-little-toy/data"
	"github.com/leihenshang/http-little-toy/msg"
	"github.com/leihenshang/http-little-toy/utils"
	"golang.org/x/net/http2"
)

var (
	helpTips = flag.Bool("h", false, "show help tips.")
	version  = flag.Bool("v", false, "show version.")
	resFile  = flag.String("resFile", "", "save result to file.")
)

func initRequestSample() *data.ToyReq {
	toyReq := &data.ToyReq{}
	flag.Var(&toyReq.Header, "header", "The http header.")
	flag.StringVar(&toyReq.Url, "u", "", "The URL you want to test.")
	flag.StringVar(&toyReq.Method, "m", http.MethodGet, "The http method.")
	flag.StringVar(&toyReq.Body, "body", "", "The http body.")
	flag.IntVar(&toyReq.Duration, "d", 10, "Duration of request.The unit is seconds.")
	flag.IntVar(&toyReq.Thread, "t", 10, "Number of threads.")
	flag.BoolVar(&toyReq.KeepAlive, "keepAlive", true, "Use keep-alive for http protocol.")
	flag.BoolVar(&toyReq.Compression, "compression", true, "Use compression for http protocol.")
	flag.IntVar(&toyReq.Timeout, "timeout", 5, "the time out to wait response.the unit is seconds.")
	flag.BoolVar(&toyReq.SkipVerify, "skipVerify", false, "TLS skipVerify.")
	flag.BoolVar(&toyReq.AllowRedirects, "allowRedirects", true, "allowRedirects.")
	flag.BoolVar(&toyReq.UseHttp2, "h2", false, "useHttp2.")
	flag.StringVar(&toyReq.ClientCert, "clientCert", "", "clientCert.")
	flag.StringVar(&toyReq.ClientKey, "clientKey", "", "clientKey.")
	flag.StringVar(&toyReq.CaCert, "caCert", "", "caCert.")
	flag.Parse()
	return toyReq
}

func main() {
	msg.SetLocalize(msg.Localize_Cn)
	toyReq := initRequestSample()
	toyReq.TipsAndHelp(*helpTips, *version)
	if err := toyReq.Check(); err != nil {
		log.Fatal(err)
	}

	outputFile, err := checkResFile()
	if err != nil {
		log.Fatal(err)
	} else if outputFile != nil {
		defer outputFile.Close()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	respChan := make(chan data.RequestStats, toyReq.Thread)

	allAggregate := data.RequestStats{MinReqTime: time.Duration(math.MaxInt64)}

	header1 := msg.MsgHeader.Sprintf(toyReq.Thread, toyReq.Duration)
	allAggregate.Res = append(allAggregate.Res, header1)
	fmt.Println(header1)

	header2 := msg.MsgSplitLine.Sprintf()
	allAggregate.Res = append(allAggregate.Res, header2)
	fmt.Println(header2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(toyReq.Duration)*time.Second)
	defer cancel()

	for i := 0; i < toyReq.Thread; i++ {
		go func() {
			client, err := genHttpClient(toyReq)
			if err != nil {
				log.Printf("genHttpClient error: %v", err)
				respChan <- data.RequestStats{ErrNum: 1}
				return
			}
			aggregate := data.RequestStats{MinReqTime: time.Duration(math.MaxInt64)}
		LOOP:
			for {
				size, d, _, err := doReq(client, toyReq)
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

	for allAggregate.RespNum < toyReq.Thread {
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
	if outputFile != nil {
		outputFile.WriteString(strings.Join(allAggregate.Res, "\n"))
	}
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

func genHttpClient(reqSample *data.ToyReq) (client *http.Client, err error) {
	client = &http.Client{}

	disableKeepAlive := !reqSample.KeepAlive
	disableCompression := !reqSample.Compression

	client.Transport = &http.Transport{
		ResponseHeaderTimeout: time.Duration(reqSample.Timeout) * time.Second,
		DisableCompression:    disableCompression,
		DisableKeepAlives:     disableKeepAlive,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: reqSample.SkipVerify},
	}

	if !reqSample.AllowRedirects {
		//returning an error when trying to redirect. This prevents the redirection from happening.
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirection not allowed")
		}
	}

	if reqSample.ClientCert == "" && reqSample.ClientKey == "" && reqSample.CaCert == "" {
		return client, nil
	}

	if reqSample.ClientCert == "" {
		return nil, fmt.Errorf("client certificate can't be empty")
	}
	if reqSample.ClientKey == "" {
		return nil, fmt.Errorf("client key can't be empty")
	}
	cert, err := tls.LoadX509KeyPair(reqSample.ClientCert, reqSample.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %v", err)
	}

	// load our CA certificate
	clientCACert, err := os.ReadFile(reqSample.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: reqSample.SkipVerify,
	}

	// 复用已配置的Transport，避免丢失超时、压缩等配置
	t := client.Transport.(*http.Transport)
	t.TLSClientConfig = tlsConfig

	if reqSample.UseHttp2 {
		if err = http2.ConfigureTransport(t); err != nil {
			return nil, fmt.Errorf("HTTP/2 configuration error: %v", err)
		}
	}

	return client, nil
}

func doReq(client *http.Client, reqObj *data.ToyReq) (respSize int, duration time.Duration, bodyBytes []byte, err error) {
	respSize = -1
	duration = -1

	req, err := http.NewRequest(reqObj.Method, reqObj.Url, strings.NewReader(reqObj.Body))
	if err != nil {
		log.Printf("new request failed, err:%v\n", err)
		return
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", data.AppName, data.Version))

	for _, v := range reqObj.Header {
		if temp := strings.SplitN(v, ":", 2); len(temp) == 2 {
			req.Header.Add(temp[0], temp[1])
		} else {
			log.Printf("split header error, value: %+v, split len: %v", v, len(temp))
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	duration = time.Since(start)

	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("an error occurred doing request(io readAll): %v", err)
		return respSize, duration, bodyBytes, fmt.Errorf("failed to read response body: %w", err)
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
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		err = fmt.Errorf("client error: %d", resp.StatusCode)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		err = fmt.Errorf("server error: %d", resp.StatusCode)
	default:
		err = fmt.Errorf("http-code: %d, header: %v, content: %s", resp.StatusCode, resp.Header, string(bodyBytes))
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

func checkResFile() (*os.File, error) {
	if *resFile == "" {
		return nil, nil
	}

	file, err := utils.CreateFile(*resFile)
	if err != nil {
		return nil, err
	}
	return file, nil
}
