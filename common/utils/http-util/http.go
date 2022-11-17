package http_util

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	checkutil "github.com/leihenshang/http-little-toy/common/utils/net_util"
)

func GetHttpMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		// http.MethodConnect,
		// http.MethodOptions,
		// http.MethodTrace,
	}
}

func CheckHttpMethod(method string) error {
	for _, v := range GetHttpMethods() {
		if v == method {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("%s is not in %s.", method, fmt.Sprintf("%v", GetHttpMethods())))
}

func CalculateHttpHeadersSize(headers http.Header) (result int64) {
	result = 0

	for k, v := range headers {
		result += int64(len(k) + len(": \r\n"))
		for _, s := range v {
			result += int64(len(s))
		}
	}

	result += int64(len("\r\n"))

	return result
}

func CheckUrlAddr(urlAddr string) (err error) {

	if urlAddr == "" {
		return errors.New("url is empty")
	}

	urlObj, urlErr := url.Parse(urlAddr)
	if urlErr != nil {
		return errors.New("an error occurred while parsing the url")
	}

	if urlObj.Scheme != "http" && urlObj.Scheme != "https" {
		return errors.New("url schema illegal:" + urlObj.Scheme)
	}

	hostName := urlObj.Hostname()
	port := urlObj.Port()

	portErr := checkutil.ConnectivityTest(fmt.Sprintf("%s:%s", hostName, port))
	if portErr != nil {
		return portErr
	}

	return
}
