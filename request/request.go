package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	httputil "tangzq/http-little-toy/common/utils/http-util"
	"tangzq/http-little-toy/model"
	"time"

	"golang.org/x/net/http2"
)

func GetHttpClient(disableKeepAlive bool,
	disableCompression bool,
	timeout time.Duration,
	skipVerify bool,
	allowRedirects bool,
	clientCert, clientKey, caCert string,
	useHttp2 bool,
) (client *http.Client, err error) {
	client = &http.Client{}

	client.Transport = &http.Transport{
		ResponseHeaderTimeout: time.Millisecond * time.Duration(timeout),
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
		return nil, fmt.Errorf("Unable to load cert tried to load %v and %v but got %v", clientCert, clientKey, err)
	}

	// Load our CA certificate
	clientCACert, err := ioutil.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("Unable to open cert %v", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: skipVerify,
	}

	tlsConfig.BuildNameToCertificate()
	t := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if useHttp2 {
		http2.ConfigureTransport(t)
	}
	client.Transport = t
	return client, nil
}

func HandleReq(_ context.Context, client *http.Client, reqObj model.Request) (respSize int, duration time.Duration, err error) {
	respSize = -1
	duration = -1

	formValues := url.Values{}
	if len(reqObj.FormData) > 0 {
		for k, v := range reqObj.FormData {
			formValues.Add(k, v)
		}
	}

	reader := strings.NewReader(formValues.Encode())
	req, err := http.NewRequest(reqObj.Method, reqObj.Url, reader)
	if err != nil {
		fmt.Printf("new request failed, err:%v\n", err)
		return
	}

	if len(reqObj.FormData) > 0 {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if reqObj.Header != nil {
		for k, v := range reqObj.Header {
			req.Header.Add(k, v)
		}
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("an error occurred doing request:", err)
	}

	headerSize := 0
	if len(resp.Header) > 0 {
		headerSize = int(httputil.CalculateHttpHeadersSize(resp.Header))
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		duration = time.Since(start)
		respSize = len(body) + headerSize
	} else if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect {
		duration = time.Since(start)
		respSize = int(resp.ContentLength) + headerSize
	} else {
		fmt.Println("received status code", resp.StatusCode, "header", resp.Header, "content", string(body))
	}

	return
}
