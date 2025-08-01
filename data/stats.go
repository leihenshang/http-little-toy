package data

import (
	"fmt"
	"time"
)

// RequestStats request statistics
type RequestStats struct {
	RespSize   int64
	Duration   time.Duration
	MinReqTime time.Duration
	MaxReqTime time.Duration
	ErrNum     int
	SuccessNum int
	RespNum    int
}

// PrintStats print statistics data
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

	fmt.Printf("number of success: %v ,number of failed: %v,read: %v KB \n", r.SuccessNum, r.ErrNum, r.RespSize/1024)
	fmt.Printf("requests/sec %.2f , transfer/sec %.2f KB, average request time: %v \n", perSecondTimes, byteRate/1024, averageRequestTime)
	fmt.Printf("the slowest request:%v \n", r.MaxReqTime)
	fmt.Printf("the fastest request:%v \n", r.MinReqTime)

}
