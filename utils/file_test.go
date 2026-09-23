package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFile(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
		err      error
	}{{
		name:     "File does not exist, directory exists",
		filename: filepath.Join(tempDir, "test1.txt"),
		wantErr:  false,
		err:      nil,
	}, {
		name:     "File does not exist, directory does not exist",
		filename: filepath.Join(tempDir, "subdir", "test2.txt"),
		wantErr:  false,
		err:      nil,
	}, {
		// L6：同名文件现在允许覆盖（O_TRUNC），不再报 os.ErrExist
		name:     "File already exists (overwritten)",
		filename: filepath.Join(tempDir, "existing.txt"),
		wantErr:  false,
		err:      nil,
	}, {
		name:     "Current directory",
		filename: "test_current.txt",
		wantErr:  false,
		err:      nil,
	}}

	// 为 "File already exists" 测试创建文件
	existingFile := filepath.Join(tempDir, "existing.txt")
	if _, err := os.Create(existingFile); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := CreateFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != tt.err {
					t.Errorf("CreateFile() error = %v, want %v", err, tt.err)
				}
				return
			}
			if file == nil {
				t.Error("CreateFile() returned nil file")
				return
			}
			file.Close()

			// 验证文件是否创建成功
			if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
				t.Errorf("CreateFile() did not create file: %s", tt.filename)
			}
		})
	}

	// 清理当前目录创建的文件
	if _, err := os.Stat("test_current.txt"); err == nil {
		os.Remove("test_current.txt")
	}
}

// TestCreateFile_OverwritesAndTruncates 验证 L6：已存在文件被覆盖且截断，
// 不留旧内容；同时验证新文件的父目录链被自动创建。
func TestCreateFile_OverwritesAndTruncates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "out.txt")

	first, err := CreateFile(path)
	if err != nil {
		t.Fatalf("CreateFile() unexpected error: %v", err)
	}
	first.WriteString("old content that is quite long......")
	first.Close()

	second, err := CreateFile(path)
	if err != nil {
		t.Fatalf("CreateFile() on existing file: unexpected error %v", err)
	}
	second.WriteString("new")
	second.Close()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q (truncation failed)", got, "new")
	}
}
