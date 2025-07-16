package serverfilemanagement

import (
	"errors"
	"testing"
)

func TestGetValidatePath(t *testing.T) {
	// 初始化 ServerFileManagement 实例
	rootPath := []ServerFileManagement_PathEntry{
		{RealPath: "D:\\aaa", VirtualPath: "rootV"},
	}
	s := ServerFileManagement{RootPath: rootPath}
	s.preprocessRoots()

	tests := []struct {
		name     string
		input    string
		expected string
		err      error
	}{
		{
			name:     "Empty Path (Special Case)",
			input:    "",
			expected: "",
			err:      nil,
		},
		{
			name:     "Valid Path Inside Root",
			input:    "D:\\aaa\\file.txt",
			expected: "D:/aaa/file.txt",
			err:      nil,
		},
		{
			name:     "Valid virtual path",
			input:    "rootV/file.txt",
			expected: "D:/aaa/file.txt",
			err:      nil,
		},
		{
			name:     "Invalid Path Outside Root",
			input:    "C:\\file.txt",
			expected: "",
			err:      Err_ServerFileManagement_InvalidPath, // 修改为路径无效错误
		},
		{
			name:     "Relative Path (Not Absolute)",
			input:    "../file.txt",
			expected: "",
			err:      Err_ServerFileManagement_NotAbsolutePath, // 使用已定义的错误
		},
		{
			name:     "Path Traversal Attempt",
			input:    "D:\\aaa\\..\\file.txt",
			expected: "",
			err:      Err_ServerFileManagement_InvalidPath, // 修改为路径无效错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.GetValidatePath(tt.input)
			if !errors.Is(err, tt.err) {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			if result != tt.expected {
				t.Errorf("expected result %v, got %v", tt.expected, result)
			}
		})
	}
}
