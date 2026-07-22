//go:build !windows

package processors

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

func (Play) openFolderAndSelectFile(filePath string) error {
	folderPath := filepath.Dir(filePath)
	if runtime.GOOS == "darwin" {
		return exec.Command("open", folderPath).Run()
	}
	return exec.Command("xdg-open", folderPath).Run()
}
