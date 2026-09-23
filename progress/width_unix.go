//go:build unix

package progress

import (
	"os"
	"syscall"
	"unsafe"
)

// winsize 为 TIOCGWINSZ 的内核交互结构体（标准库 syscall 未导出）。
// winsize mirrors the kernel struct used by TIOCGWINSZ.
type winsize struct {
	row, col, xpixel, ypixel uint16
}

// terminalWidth 通过 ioctl(TIOCGWINSZ) 查询终端列宽，失败返回 0。
// 仅使用标准库 syscall，CGO_ENABLED=0 下交叉编译不受影响；
// Linux 各架构的 TIOCGWINSZ 值由 asm-generic 统一约定。
// terminalWidth queries the console columns via ioctl(TIOCGWINSZ)
// using only the standard library.
func terminalWidth(f *os.File) int {
	if f == nil {
		return 0
	}
	var ws winsize
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 || ws.col <= 0 {
		return 0
	}
	return int(ws.col)
}
