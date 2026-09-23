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
	"github.com/leihenshang/http-little-toy/progress"
	"github.com/leihenshang/http-little-toy/utils"
	"golang.org/x/net/http2"
)

var (
	helpTips   = false
	version    = false
	resFile    = ""
	format     = "raw"
	outputLang = "en"
	progressOn = true
	toyReq     = &data.ToyReq{}
)

// maxErrBodyBytes 错误响应的 body 采样上限（字节）。仅采样用于诊断信息，
// 其余边读边丢弃计数，避免高并发下整块响应体常驻内存（P1-1）。
const maxErrBodyBytes = 512

func initParameters() {
	flag.BoolVar(&helpTips, "h", false, "show help tips.")
	flag.BoolVar(&version, "v", false, "show version.")
	flag.StringVar(&resFile, "resFile", "", "save result to file.")
	flag.StringVar(&format, "format", "raw", "output format (json/csv/raw).")
	flag.StringVar(&outputLang, "lang", "en", "output language (en/zh).")
	flag.BoolVar(&progressOn, "progress", true, "show progress bar (only in raw format on an interactive terminal).")

	flag.Var(&toyReq.Header, "header", "The http header.")
	flag.StringVar(&toyReq.Url, "u", "", "The URL you want to test.")
	flag.StringVar(&toyReq.Method, "m", http.MethodGet, "The http method.")
	flag.StringVar(&toyReq.Body, "body", "", "The http body.")
	flag.IntVar(&toyReq.Duration, "d", 10, "Duration of request.The unit is seconds.")
	flag.IntVar(&toyReq.Thread, "t", 10, "Number of threads.")
	flag.BoolVar(&toyReq.KeepAlive, "keepAlive", true, "Use keep-alive for http protocol.")
	flag.BoolVar(&toyReq.Compression, "compression", true, "Use compression for http protocol.")
	flag.IntVar(&toyReq.Timeout, "timeout", 10, "the time out to wait response.the unit is seconds.")
	flag.BoolVar(&toyReq.SkipVerify, "skipVerify", false, "TLS skipVerify.")
	flag.BoolVar(&toyReq.AllowRedirects, "allowRedirects", true, "allowRedirects.")
	flag.BoolVar(&toyReq.UseHttp2, "h2", false, "useHttp2.")
	flag.StringVar(&toyReq.ClientCert, "clientCert", "", "clientCert.")
	flag.StringVar(&toyReq.ClientKey, "clientKey", "", "clientKey.")
	flag.StringVar(&toyReq.CaCert, "caCert", "", "caCert.")
	flag.Parse()

	toyReq.TipsAndHelp(helpTips, version)

	// L3：显示类参数在白名单内取值，非法值快速失败
	switch format {
	case "raw", "json", "csv":
	default:
		log.Fatalf("invalid -format %q, valid values: raw/json/csv.", format)
	}
	switch outputLang {
	case "en", "zh":
	default:
		log.Fatalf("invalid -lang %q, valid values: en/zh.", outputLang)
	}

	msg.SetLocalize(msg.Localize(outputLang))
}

func main() {
	initParameters()

	if err := toyReq.Validate(); err != nil {
		log.Fatal(err)
	}

	// S1：显式禁用 TLS 校验时打印一次告警（stderr，不污染数据输出）
	if toyReq.SkipVerify {
		fmt.Fprintln(os.Stderr, msg.MsgWarnSkipVerify.Sprintf())
	}

	// S2：自定义头部在启动期一次性解析+净化（CR/LF 拒绝），worker 热路径只做 Add
	customHeaders, err := toyReq.ParseHeaders()
	if err != nil {
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

	allAggregate := data.RequestStats{MinReqTime: time.Duration(math.MaxInt64), Format: format, Url: toyReq.Url}

	header1 := msg.MsgHeader.Sprintf(toyReq.Thread, toyReq.Duration)
	allAggregate.Res = append(allAggregate.Res, header1)
	printLByFormat(format, header1)

	header2 := msg.MsgSplitLine.Sprintf()
	allAggregate.Res = append(allAggregate.Res, header2)
	printLByFormat(format, header2)

	client, err := genHttpClient(toyReq)
	if err != nil {
		log.Fatalf("genHttpClient error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(toyReq.Duration)*time.Second)
	defer cancel()

	// 进度条：worker 热路径只做原子计数，由独立协程按 tick 刷新渲染。
	// Real-time counters are atomic; a dedicated goroutine renders the bar.
	tracker := progress.NewTracker()
	if progress.ShouldRender(progressOn, format, os.Stdout) {
		go progress.Render(ctx, tracker, os.Stdout, time.Duration(toyReq.Duration)*time.Second, progress.DefaultInterval,
			func() int { return progress.Width(os.Stdout) })
	}

	testStart := time.Now()
	startWorkers(ctx, client, toyReq, customHeaders, tracker, respChan)
	collectResults(respChan, signalChan, cancel, toyReq.Thread, &allAggregate)
	allAggregate.Elapsed = time.Since(testStart)

	allAggregate.PrintStats()

	if summary := tracker.ErrorSummary(); summary != "" {
		if format == "raw" {
			allAggregate.Res = append(allAggregate.Res, summary)
			printLByFormat(format, summary)
		} else {
			// json/csv 的 -resFile 保持纯结构化内容，错误摘要走 stderr
			fmt.Fprintln(os.Stderr, summary)
		}
	}

	if outputFile != nil {
		outputFile.WriteString(strings.Join(allAggregate.Res, "\n"))
	}
}

// startWorkers 启动 toyReq.Thread 个压测协程，各自聚合统计并写入 out。
// Q3：从 main() 拆出，main 只保留装配与调度。
func startWorkers(ctx context.Context, client *http.Client, reqObj *data.ToyReq, headers [][2]string, tracker *progress.Tracker, out chan<- data.RequestStats) {
	for i := 0; i < reqObj.Thread; i++ {
		go func() {
			aggregate := data.RequestStats{MinReqTime: time.Duration(math.MaxInt64)}
		LOOP:
			for {
				size, d, err := doReq(client, reqObj, headers)
				if size > 0 && err == nil {
					aggregate.Duration += d
					aggregate.SuccessNum++
					aggregate.MaxReqTime = maxTime(aggregate.MaxReqTime, d)
					aggregate.MinReqTime = minTime(aggregate.MinReqTime, d)
					aggregate.RespSize += int64(size)
					tracker.OnSuccess(size)
				} else {
					// 错误不再逐条打印（避免高并发下刷屏与锁争用，
					// 也会打乱进度条），由 tracker 聚合后结束时输出摘要。
					aggregate.ErrNum++
					tracker.OnFailure(err)
				}

				select {
				case <-ctx.Done():
					break LOOP
				default:
				}
			}
			out <- aggregate
		}()
	}
}

// collectResults 汇总各 worker 统计；收到 Ctrl+C 信号则提前终止计时。
func collectResults(in <-chan data.RequestStats, signals <-chan os.Signal, cancel context.CancelFunc, thread int, dst *data.RequestStats) {
	for dst.RespNum < thread {
		select {
		case r := <-in:
			dst.ErrNum += r.ErrNum
			dst.SuccessNum += r.SuccessNum
			dst.RespSize += r.RespSize
			dst.Duration += r.Duration
			dst.MinReqTime = minTime(dst.MinReqTime, r.MinReqTime)
			dst.MaxReqTime = maxTime(dst.MaxReqTime, r.MaxReqTime)
			dst.RespNum++
		case <-signals:
			cancel()
		}
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
	disableKeepAlive := !reqSample.KeepAlive
	disableCompression := !reqSample.Compression

	// 空闲连接超时设置为测试持续时间的1.5倍，确保测试期间连接可用
	idleConnTimeout := max(time.Duration(reqSample.Duration)*time.Second*3/2, 30*time.Second)

	t := &http.Transport{
		ResponseHeaderTimeout: time.Duration(reqSample.Timeout) * time.Second,
		DisableCompression:    disableCompression,
		DisableKeepAlives:     disableKeepAlive,
		IdleConnTimeout:       idleConnTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: reqSample.SkipVerify},
		ForceAttemptHTTP2:     reqSample.UseHttp2, // L1：自定义 TLSClientConfig 时隐式 h2 失效，需显式开关
		MaxIdleConns:          reqSample.Thread,   // 增加最大空闲连接数
		MaxIdleConnsPerHost:   reqSample.Thread,   // 增加每个主机的最大空闲连接数
		MaxConnsPerHost:       reqSample.Thread,   // 增加每个主机的最大连接数
	}

	client = &http.Client{
		Timeout:   time.Duration(reqSample.Timeout) * time.Second,
		Transport: t,
	}

	if !reqSample.AllowRedirects {
		//returning an error when trying to redirect. This prevents the redirection from happening.
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirection not allowed")
		}
	}

	// mTLS：任一证书参数出现，则校验并加载（要求三者齐备）
	if reqSample.ClientCert != "" || reqSample.ClientKey != "" || reqSample.CaCert != "" {
		tlsConfig, tlsErr := buildMutualTLSConfig(reqSample)
		if tlsErr != nil {
			return nil, tlsErr
		}
		t.TLSClientConfig = tlsConfig
	}

	// L1：HTTP/2 配置与证书分支解耦，任何场景下 -h2 都生效
	if reqSample.UseHttp2 {
		if err = http2.ConfigureTransport(t); err != nil {
			return nil, fmt.Errorf("HTTP/2 configuration error: %w", err)
		}
	}

	return client, nil
}

// buildMutualTLSConfig 加载客户端证书、私钥与 CA 证书，构建双向 TLS 配置。
func buildMutualTLSConfig(reqSample *data.ToyReq) (*tls.Config, error) {
	if reqSample.ClientCert == "" {
		return nil, fmt.Errorf("client certificate can't be empty")
	}
	if reqSample.ClientKey == "" {
		return nil, fmt.Errorf("client key can't be empty")
	}
	if reqSample.CaCert == "" {
		return nil, fmt.Errorf("CA certificate can't be empty")
	}

	cert, err := tls.LoadX509KeyPair(reqSample.ClientCert, reqSample.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// load our CA certificate
	clientCACert, err := os.ReadFile(reqSample.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: reqSample.SkipVerify,
	}, nil
}

// doReq 发送单次请求。返回响应字节数（含头部估算）、耗时与错误。
// headers 必须是启动期经 ToyReq.ParseHeaders() 净化后的键值对（S2）。
func doReq(client *http.Client, reqObj *data.ToyReq, headers [][2]string) (respSize int, duration time.Duration, err error) {
	req, err := http.NewRequest(reqObj.Method, reqObj.Url, strings.NewReader(reqObj.Body))
	if err != nil {
		return -1, 0, fmt.Errorf("new request failed: %w", err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", data.AppName, data.Version))
	for _, h := range headers {
		req.Header.Add(h[0], h[1])
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration = time.Since(start)
	if err != nil {
		// CheckRedirect 拦截重定向等场景下 resp 非 nil，需关闭防止连接泄露
		if resp != nil {
			resp.Body.Close()
		}
		return -1, duration, err
	}
	defer resp.Body.Close()

	// P1-1：不整块读入 body。采样前 maxErrBodyBytes 字节用于错误诊断，
	// 其余 io.Copy 到 Discard 只计数，避免大响应体高并发下的内存压力。
	sample, err := io.ReadAll(io.LimitReader(resp.Body, maxErrBodyBytes))
	if err != nil {
		return -1, duration, fmt.Errorf("failed to read response body: %w", err)
	}
	rest, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return -1, duration, fmt.Errorf("failed to read response body: %w", err)
	}
	bodySize := int64(len(sample)) + rest
	headerSize := int(calculateHttpHeadersSize(resp.Header))

	// L2：按状态码区间分类，全部 2xx 视为成功；
	// 3xx 未被客户端跟随时计为失败；4xx/5xx 归类为客户端/服务端错误。
	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		respSize = int(bodySize) + headerSize
	case code >= 300 && code < 400:
		err = fmt.Errorf("redirect not followed: %d", code)
	case code >= 400 && code < 500:
		err = fmt.Errorf("client error: %d, body: %s", code, sample)
	case code >= 500:
		err = fmt.Errorf("server error: %d, body: %s", code, sample)
	default:
		err = fmt.Errorf("unexpected status: %d, body: %s", code, sample)
	}

	return respSize, duration, err
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
	if resFile == "" {
		return nil, nil
	}

	file, err := utils.CreateFile(resFile)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func printLByFormat(format string, msg string) {
	if format == "raw" {
		fmt.Println(msg)
	}
}
