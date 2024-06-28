package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	checkutil "github.com/leihenshang/http-little-toy/common/utils/net"
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

func CheckUrl(urlAddr string) (err error) {
	url, urlErr := url.Parse(urlAddr)
	if urlErr != nil {
		return errors.New("an error occurred while parsing the url," + urlErr.Error())
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return errors.New("url schema illegal:" + url.Scheme)
	}

	portErr := checkutil.ConnectivityTest(fmt.Sprintf("%s:%s", url.Hostname(), url.Port()))
	if portErr != nil {
		return portErr
	}

	return
}
