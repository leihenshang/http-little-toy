package data

import (
	"fmt"
	"time"

	"github.com/leihenshang/http-little-toy/msg"
)

type RequestStats struct {
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

	perSecondTimes := float64(r.SuccessNum) / averageThreadDuration.Seconds()
	byteRate := float64(r.RespSize) / averageThreadDuration.Seconds()
	res := msg.MsgStats.Sprintf(r.SuccessNum, r.ErrNum, r.RespSize/1024,
		perSecondTimes, byteRate/1024, averageRequestTime, r.MaxReqTime, r.MinReqTime)
	r.Res = append(r.Res, res)
	fmt.Println(res)
}
