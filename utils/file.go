package utils

import (
	"os"
	"path/filepath"
)

func CreateFile(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); err == nil {
		return nil, os.ErrExist
	}
	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}
