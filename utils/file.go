package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateFile 创建或覆盖写入目标文件（L6）。
// 使用 O_CREATE|O_TRUNC 原子处理"已存在"场景：既允许覆盖输出，
// 也消除了原先 stat→create 之间的 TOCTOU 竞态（S3）。
// CreateFile creates or truncates the target file atomically, allowing
// overwrite and removing the previous stat→create TOCTOU race.
func CreateFile(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", filename, err)
	}
	return file, nil
}
