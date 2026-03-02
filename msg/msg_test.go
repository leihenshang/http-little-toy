package msg

import (
	"testing"
)

func TestSetLocalize(t *testing.T) {
	tests := []struct {
		name     string
		localize Localize
	}{
		{
			name:     "Set English localization",
			localize: Localize_En,
		},
		{
			name:     "Set Chinese localization",
			localize: Localize_Cn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLocalize(tt.localize)
			// 验证本地化设置是否生效（通过后续的 Sprintf 测试）
		})
	}
}

func TestToyMsg_Sprintf(t *testing.T) {
	tests := []struct {
		name     string
		localize Localize
		msg      *ToyMsg
		args     []interface{}
		expected string
	}{
		{
			name:     "English header message",
			localize: Localize_En,
			msg:      &MsgHeader,
			args:     []interface{}{10, 30},
			expected: "use 10 coroutines,duration 30 seconds.",
		},
		{
			name:     "Chinese header message",
			localize: Localize_Cn,
			msg:      &MsgHeader,
			args:     []interface{}{10, 30},
			expected: "使用 [10] 个协程，持续 [30] 秒",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLocalize(tt.localize)
			result := tt.msg.Sprintf(tt.args...)
			if result != tt.expected {
				t.Errorf("ToyMsg.Sprintf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToyMsg_Printf(t *testing.T) {
	tests := []struct {
		name     string
		localize Localize
		msg      *ToyMsg
		args     []interface{}
	}{
		{
			name:     "English printf",
			localize: Localize_En,
			msg:      &MsgHeader,
			args:     []interface{}{5, 10},
		},
		{
			name:     "Chinese printf",
			localize: Localize_Cn,
			msg:      &MsgHeader,
			args:     []interface{}{5, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLocalize(tt.localize)
			// 测试 Printf 方法是否能正常执行而不 panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ToyMsg.Printf() panicked: %v", r)
				}
			}()
			tt.msg.Printf(tt.args...)
		})
	}
}