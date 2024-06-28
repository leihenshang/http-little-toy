package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/leihenshang/http-little-toy/data"

	"golang.org/x/net/http2"
)

func GetHttpClient(
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
		ResponseHeaderTimeout: time.Millisecond * timeout,
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

func HandleReq(_ context.Context, client *http.Client, reqObj data.Request,
) (respSize int, duration time.Duration, bodyBytes []byte, err error) {
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

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		respSize = len(bodyBytes) + headerSize
	} else if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect {
		respSize = int(resp.ContentLength) + headerSize
	} else {
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
