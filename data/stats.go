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

	// Elapsed 测试实际挂钟时长，由调用方在聚合完成后设置；
	// 吞吐（requests/sec、transfer/sec）以它为分母，含失败请求，
	// 目标宕机时仍能反映真实发送速率。未设置时回退用线程忙碌时长。
	Elapsed time.Duration
}

// statsCSVHeader CSV 表头（L5）/ CSV header row
const statsCSVHeader = "url,threads,success,failed,readKB,requestsPerSec,transferKBps,avgRequestMs,maxRequestMs,minRequestMs,elapsedSeconds"

// statsJSON 面向输出的统计快照（Q2）：时长用毫秒浮点、体积用 KB 浮点，
// 避免直接 Marshal 时 time.Duration 被序列化成纳秒整数、整数除法丢精度。
// statsJSON is the output-friendly snapshot: durations in ms, sizes in KB.
type statsJSON struct {
	Url            string  `json:"url"`
	Threads        int     `json:"threads"`
	Success        int     `json:"success"`
	Failed         int     `json:"failed"`
	ReadKB         float64 `json:"readKB"`
	RequestsPerSec float64 `json:"requestsPerSec"`
	TransferKBps   float64 `json:"transferKBps"`
	AvgRequestMs   float64 `json:"avgRequestMs"`
	MaxRequestMs   float64 `json:"maxRequestMs"`
	MinRequestMs   float64 `json:"minRequestMs"`
	ElapsedSeconds float64 `json:"elapsedSeconds"`
}

func (r *RequestStats) PrintStats() {
	// 平均每个线程的忙碌时长 ≈ 实际测试时长
	avgThreadBusy := time.Duration(0)
	if r.RespNum > 0 {
		avgThreadBusy = r.Duration / time.Duration(r.RespNum)
	}

	avgRequestTime := time.Duration(0)
	if r.SuccessNum > 0 {
		avgRequestTime = r.Duration / time.Duration(r.SuccessNum)
	}

	// 吞吐分母：优先用测试实际挂钟时长（Elapsed），未设置时回退线程忙碌均时长；
	// requestsPerSec 按总请求数（成功+失败）计算。
	denom := r.Elapsed
	if denom <= 0 {
		denom = avgThreadBusy
	}

	var perSecondTimes, byteRate float64
	if sec := denom.Seconds(); sec > 0 {
		perSecondTimes = float64(r.SuccessNum+r.ErrNum) / sec
		byteRate = float64(r.RespSize) / sec
	}

	// L4：全部失败时 MinReqTime 仍是初始化哨兵值 MaxInt64，兜底为 0，
	// 避免打印出 2562047h... 的天文数字。
	minReqTime := r.MinReqTime
	if r.SuccessNum == 0 {
		minReqTime = 0
	}

	readKB := float64(r.RespSize) / 1024
	text := msg.MsgStats.Sprintf(r.SuccessNum, r.ErrNum, readKB,
		perSecondTimes, byteRate/1024, avgRequestTime, r.MaxReqTime, minReqTime)
	r.Res = append(r.Res, text)

	switch r.Format {
	case "json":
		payload := statsJSON{
			Url:            r.Url,
			Threads:        r.RespNum,
			Success:        r.SuccessNum,
			Failed:         r.ErrNum,
			ReadKB:         readKB,
			RequestsPerSec: perSecondTimes,
			TransferKBps:   byteRate / 1024,
			AvgRequestMs:   toMS(avgRequestTime),
			MaxRequestMs:   toMS(r.MaxReqTime),
			MinRequestMs:   toMS(minReqTime),
			ElapsedSeconds: denom.Seconds(),
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Error marshalling stats to JSON:", err)
			return
		}
		r.Res = r.Res[0:0]
		r.Res = append(r.Res, string(jsonBytes))
		fmt.Println(string(jsonBytes))
	case "csv":
		// L5：实现合法 CSV（表头 + 一行数据），url 含逗号时由 %q 带引号转义
		csv := statsCSVHeader + "\n" + fmt.Sprintf("%q,%d,%d,%d,%.2f,%.2f,%.2f,%.3f,%.3f,%.3f,%.3f\n",
			r.Url, r.RespNum, r.SuccessNum, r.ErrNum, readKB,
			perSecondTimes, byteRate/1024,
			toMS(avgRequestTime), toMS(r.MaxReqTime), toMS(minReqTime), denom.Seconds())
		r.Res = r.Res[0:0]
		r.Res = append(r.Res, csv)
		fmt.Print(csv)
	default:
		fmt.Println(text)
	}
}

// toMS 时长转毫秒浮点 / converts duration to milliseconds
func toMS(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
