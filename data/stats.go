package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/leihenshang/http-little-toy/msg"
)

type RequestStats struct {
	Url        string
	Format     string
	RespSize   int64
	Duration   time.Duration
	MinReqTime time.Duration
	MaxReqTime time.Duration
	ErrNum     int
	SuccessNum int
	RespNum    int
	Res        []string
}

func (r *RequestStats) PrintStats() {
	averageThreadDuration := func() time.Duration {
		if time.Duration(r.RespNum) <= 0 {
			return 0
		}
		return r.Duration / time.Duration(r.RespNum)
	}()

	averageRequestTime := func() time.Duration {
		if time.Duration(r.SuccessNum) <= 0 {
			return 0
		}
		return r.Duration / time.Duration(r.SuccessNum)
	}()

	var perSecondTimes float64
	var byteRate float64
	if averageThreadDuration.Seconds() > 0 {
		perSecondTimes = float64(r.SuccessNum) / averageThreadDuration.Seconds()
		byteRate = float64(r.RespSize) / averageThreadDuration.Seconds()
	} else {
		perSecondTimes = 0
		byteRate = 0
	}
	res := msg.MsgStats.Sprintf(r.SuccessNum, r.ErrNum, r.RespSize/1024,
		perSecondTimes, byteRate/1024, averageRequestTime, r.MaxReqTime, r.MinReqTime)
	r.Res = append(r.Res, res)
	if r.Format == "json" {
		jsonBytes, err := json.Marshal(r)
		if err != nil {
			fmt.Println("Error marshalling stats to JSON:", err)
			return
		}
		res = string(jsonBytes)
		r.Res = r.Res[0:0]
		r.Res = append(r.Res, res)
		return
	}
	fmt.Println(res)
}
