package tray

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser 在默认浏览器中打开网页
func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("无法打开浏览器: %v\n", err)
	}
}
