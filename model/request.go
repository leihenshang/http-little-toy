package model

import (
	httputil "leihenshang/http-little-toy/common/utils/http-util"
	"time"
)

type Request struct {
	Url    string   `json:"url"`
	Body   string   `json:"body"`
	Method string   `json:"method"`
	Header []string `json:"header"`
}

type RequestStats struct {
	RespSize   int64
	Duration   time.Duration
	MinReqTime time.Duration
	MaxReqTime time.Duration
	ErrNum     int
	SuccessNum int
}

func (r Request) Valid() (err error) {
	// 检查 url 格式
	if urlErr := httputil.CheckUrlAddr(r.Url); urlErr != nil {
		return urlErr
	}

	// 检查 method
	if methodErr := httputil.CheckHttpMethod(r.Method); methodErr != nil {
		return methodErr
	}

	return
}
