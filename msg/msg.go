package msg

import (
	"fmt"
)

type Localize string

const (
	Localize_En Localize = "en"
	Localize_Cn Localize = "cn"
)

var (
	localize Localize = Localize_En
)

func SetLocalize(l Localize) {
	switch l {
	case Localize_Cn:
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
var MsgStats ToyMsg = ToyMsg{
	`成功: %v ,失败: %v,读取: %v KB 
每秒请求: %.2f , 每秒传输: %.2f KB, 平均请求时间: %v 
最慢的请求:%v 
最快的请求:%v 
	`,
	`number of success: %v ,number of failed: %v,read: %v KB 
requests/sec %.2f , transfer/sec %.2f KB, average request time: %v 
the slowest request:%v 
the fastest request:%v 
	 `,
}

func (t *ToyMsg) Sprintf(args ...any) string {
	if localize == Localize_Cn {
		return fmt.Sprintf(t.Cn, args...)
	}
	return fmt.Sprintf(t.En, args...)
}

func (t *ToyMsg) Printf(args ...any) {
	if localize == Localize_Cn {
		fmt.Printf(t.Cn, args...)
		return
	}
	fmt.Printf(t.En, args...)
}
