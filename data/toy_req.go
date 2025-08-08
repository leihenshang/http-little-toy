package data

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type ToyReq struct {
	// Url target url
	Url string `json:"url"`

	// Header is  HTTP header
	Header MyStrSlice `json:"header"`

	// Body request body
	Body string `json:"body"`

	// Method http Method
	Method string `json:"method"`

	// Duration time
	Duration int `json:"duration"`

	// Thread is the number of threads
	Thread int `json:"thread"`

	// KeepAlive is whether to use keep-alive
	KeepAlive bool `json:"keepAlive"`

	// Compression
	Compression bool `json:"compression"`

	// Timeout
	Timeout int `json:"timeout"`

	// SkipVerify is whether to skip TLS verification
	SkipVerify bool `json:"skipVerify"`

	// AllowRedirects
	AllowRedirects bool `json:"allowRedirects"`

	// UseHttp2
	UseHttp2 bool `json:"useHttp2"`

	// ClientCert
	ClientCert string `json:"clientCert"`

	// ClientKey
	ClientKey string `json:"clientKey"`

	// CaCert
	CaCert string `json:"caCert"`
}

func (r *ToyReq) Check() (err error) {
	if r.Url == "" {
		err = errors.New("the URL cannot be empty.Use the \"-u\" or \"-f\" parameter to set the URL")
		return
	}

	if methodErr := checkHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}
	return nil
}

func (r *ToyReq) PrintDefault(appName string) {
	fmt.Printf("Usage: %s <options>", appName)
	fmt.Println("Options:")
	flag.VisitAll(func(flag *flag.Flag) {
		fmt.Println("\t-"+flag.Name, "\t\n\t\t", flag.Usage, "--default="+func() string {
			if flag.DefValue == "" {
				return "\"\""
			}

			return flag.DefValue
		}()+".")
	})
}

func (r *ToyReq) TipsAndHelp(helpTips, version bool) {
	if helpTips {
		r.PrintDefault(AppName)
		os.Exit(0)
	}

	if version {
		fmt.Printf("%s v%s \n", AppName, Version)
		os.Exit(0)
	}
}

type MyStrSlice []string

func (s *MyStrSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *MyStrSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func checkHttpMethod(method string) error {
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
	if _, ok := httpMethodMap[method]; ok {
		return nil
	}
	return fmt.Errorf("%s is not in %s.", method, fmt.Sprintf("%v", httpMethodMap))
}
