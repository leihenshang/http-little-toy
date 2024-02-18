package http_util

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	checkutil "github.com/leihenshang/http-little-toy/common/utils/net-util"
)

var httpMethodMap = map[string]struct{}{
	http.MethodGet:    {},
	http.MethodHead:   {},
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
	// http.MethodConnect,
	// http.MethodOptions,
	// http.MethodTrace,
}

func CheckHttpMethod(method string) error {
	if _, ok := httpMethodMap[method]; ok {
		return nil
	}
	return errors.New(fmt.Sprintf("%s is not in %s.", method, fmt.Sprintf("%v", httpMethodMap)))
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

func CheckUrl(urlAddr string) (err error) {
	urlObj, urlErr := url.Parse(urlAddr)
	if urlErr != nil {
		return errors.New("an error occurred while parsing the url," + urlErr.Error())
	}

	if urlObj.Scheme != "http" && urlObj.Scheme != "https" {
		return errors.New("url schema illegal:" + urlObj.Scheme)
	}

	portErr := checkutil.ConnectivityTest(fmt.Sprintf("%s:%s", urlObj.Hostname(), urlObj.Port()))
	if portErr != nil {
		return portErr
	}

	return
}
