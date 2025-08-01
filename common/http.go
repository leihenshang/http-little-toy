package common

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
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

func ConnectivityTest(ipPorts string) (err error) {
	netRes, err := net.DialTimeout("tcp", ipPorts, time.Second*5)
	if err != nil {
		return err
	}
	if netRes == nil {
		return fmt.Errorf("the %s is disabled.", ipPorts)
	}

	_ = netRes.Close()
	return
}

func CheckHttpMethod(method string) error {
	if _, ok := httpMethodMap[method]; ok {
		return nil
	}
	return fmt.Errorf("%s is not in %s.", method, fmt.Sprintf("%v", httpMethodMap))
}

func CheckUrl(urlAddr string) (err error) {
	url, urlErr := url.Parse(urlAddr)
	if urlErr != nil {
		return errors.New("an error occurred while parsing the url," + urlErr.Error())
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return errors.New("url schema illegal:" + url.Scheme)
	}

	portErr := ConnectivityTest(fmt.Sprintf("%s:%s", url.Hostname(), url.Port()))
	if portErr != nil {
		return portErr
	}

	return
}
