//go:build !unix && !windows

package progress

import "os"

// terminalWidth 其他平台（plan9/js/wasm 等）不探测，始终返回 0 走回退宽度。
// terminalWidth is unavailable on other platforms; always falls back.
func terminalWidth(f *os.File) int {
	return 0
}
