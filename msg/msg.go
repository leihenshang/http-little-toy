package msg

import (
	"fmt"
)

type Localize string

const (
	Localize_En Localize = "en"
	Localize_Zh Localize = "zh"
)

var (
	localize Localize = Localize_En
)

func SetLocalize(l Localize) {
	switch l {
	case Localize_Zh:
		localize = l
	case Localize_En:
		localize = l
	default:
		localize = Localize_En
	}
}

type ToyMsg struct {
	Cn string
	En string
}

var MsgHeader ToyMsg = ToyMsg{"使用 [%d] 个协程，持续 [%d] 秒", "use %d coroutines,duration %d seconds."}
var MsgSplitLine ToyMsg = ToyMsg{"---------------统计---------------", "---------------stats---------------"}

// MsgWarnSkipVerify 禁用 TLS 校验时的安全告警（S1）
// MsgWarnSkipVerify is the security warning when TLS verification is disabled.
var MsgWarnSkipVerify ToyMsg = ToyMsg{
	"警告：已通过 skipVerify=true 禁用 TLS 证书校验，存在中间人攻击风险，仅建议在测试环境使用。",
	"WARNING: TLS certificate verification is disabled (skipVerify=true); vulnerable to man-in-the-middle attacks, use in test environments only.",
}

var MsgStats ToyMsg = ToyMsg{
	`成功: %v ,失败: %v,读取: %.2f KB 
每秒请求: %.2f , 每秒传输: %.2f KB, 平均请求时间: %v 
最慢的请求:%v 
最快的请求:%v 
	`,
	`number of success: %v ,number of failed: %v,read: %.2f KB 
requests/sec %.2f , transfer/sec %.2f KB, average request time: %v 
the slowest request:%v 
the fastest request:%v 
	 `,
}

func (t *ToyMsg) Sprintf(args ...any) string {
	if localize == Localize_Zh {
		return fmt.Sprintf(t.Cn, args...)
	}
	return fmt.Sprintf(t.En, args...)
}

func (t *ToyMsg) Printf(args ...any) {
	if localize == Localize_Zh {
		fmt.Printf(t.Cn, args...)
		return
	}
	fmt.Printf(t.En, args...)
}
