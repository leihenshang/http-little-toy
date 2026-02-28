package data

import (
	"testing"
	"time"
)

func TestRequestStats_PrintStats(t *testing.T) {
	tests := []struct {
		name string
		stats *RequestStats
	}{
		{
			name: "Empty stats",
			stats: &RequestStats{
				Format: "raw",
				Url: "http://example.com",
			},
		},
		{
			name: "Stats with data",
			stats: &RequestStats{
				Format: "raw",
				Url: "http://example.com",
				RespSize: 1024,
				Duration: time.Second * 5,
				MinReqTime: time.Millisecond * 100,
				MaxReqTime: time.Millisecond * 500,
				ErrNum: 1,
				SuccessNum: 10,
				RespNum: 10,
			},
		},
		{
			name: "JSON format stats",
			stats: &RequestStats{
				Format: "json",
				Url: "http://example.com",
				RespSize: 1024,
				Duration: time.Second * 5,
				MinReqTime: time.Millisecond * 100,
				MaxReqTime: time.Millisecond * 500,
				ErrNum: 1,
				SuccessNum: 10,
				RespNum: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 PrintStats 方法是否能正常执行而不 panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RequestStats.PrintStats() panicked: %v", r)
				}
			}()
			tt.stats.PrintStats()
		})
	}
}

func TestRequestStats_PrintStats_DivisionByZero(t *testing.T) {
	// 测试除零情况
	stats := &RequestStats{
		Format: "raw",
		Url: "http://example.com",
		RespSize: 1024,
		Duration: 0, // 零持续时间
		MinReqTime: time.Millisecond * 100,
		MaxReqTime: time.Millisecond * 500,
		ErrNum: 0,
		SuccessNum: 0,
		RespNum: 0,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RequestStats.PrintStats() panicked with division by zero: %v", r)
		}
	}()

	stats.PrintStats()
}