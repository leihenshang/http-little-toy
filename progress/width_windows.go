//go:build windows

package progress

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

// CONSOLE_SCREEN_BUFFER_INFO 的等价布局（全 2 字节对齐成员，无 padding）。
type coord struct {
	x, y int16
}

type smallRect struct {
	left, top, right, bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

// terminalWidth 通过 kernel32!GetConsoleScreenBufferInfo 查询控制台窗口列宽，
// 失败（非控制台句柄等）返回 0。仅使用标准库 syscall，无第三方依赖。
// terminalWidth queries the console window width via kernel32 using only
// the standard library; returns 0 on failure or non-console handles.
func terminalWidth(f *os.File) int {
	if f == nil {
		return 0
	}
	var info consoleScreenBufferInfo
	r, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(f.Fd()), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 0
	}
	cols := int(info.window.right-info.window.left) + 1
	if cols <= 0 {
		return 0
	}
	return cols
}
