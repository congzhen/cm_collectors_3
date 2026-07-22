//go:build windows

package processors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	coinitApartmentThreaded = 0x2
	coinitDisableOLE1DDE    = 0x4
	rpcEChangedMode         = 0x80010106
	swRestore               = 9
	swpNoSize               = 0x0001
	swpNoMove               = 0x0002
	swpShowWindow           = 0x0040
)

type explorerWindow struct {
	handle uintptr
	title  string
}

type explorerWindowEnumeration struct {
	windows []explorerWindow
}

var (
	shell32                        = windows.NewLazySystemDLL("shell32.dll")
	ole32                          = windows.NewLazySystemDLL("ole32.dll")
	procSHParseDisplayName         = shell32.NewProc("SHParseDisplayName")
	procSHOpenFolderAndSelectItems = shell32.NewProc("SHOpenFolderAndSelectItems")
	procILFindLastID               = shell32.NewProc("ILFindLastID")
	procCoTaskMemFree              = ole32.NewProc("CoTaskMemFree")
	procCoInitializeEx             = ole32.NewProc("CoInitializeEx")
	procCoUninitialize             = ole32.NewProc("CoUninitialize")
	user32                         = windows.NewLazySystemDLL("user32.dll")
	kernel32                       = windows.NewLazySystemDLL("kernel32.dll")
	procEnumWindows                = user32.NewProc("EnumWindows")
	procGetClassNameW              = user32.NewProc("GetClassNameW")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procShowWindowAsync            = user32.NewProc("ShowWindowAsync")
	procSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	procBringWindowToTop           = user32.NewProc("BringWindowToTop")
	procSetWindowPos               = user32.NewProc("SetWindowPos")
	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID   = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput          = user32.NewProc("AttachThreadInput")
	procGetCurrentThreadID         = kernel32.NewProc("GetCurrentThreadId")
	enumExplorerWindowsCallback    = windows.NewCallback(enumExplorerWindowCallback)
)

// openFolderAndSelectFile 使用 Windows Shell API 打开父文件夹并选中目标文件。
// 该调用将窗口交给系统 Shell 管理，不创建需要服务端等待或终止的 explorer.exe 子进程。
func (Play) openFolderAndSelectFile(filePath string) error {
	filePathUTF16, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return fmt.Errorf("转换文件路径失败: %w", err)
	}
	folderPath := filepath.Dir(filePath)
	folderPathUTF16, err := windows.UTF16PtrFromString(folderPath)
	if err != nil {
		return fmt.Errorf("转换文件夹路径失败: %w", err)
	}

	// COM 初始化和 Shell API 调用必须发生在同一个系统线程。
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	initializeResult, _, _ := procCoInitializeEx.Call(
		0,
		uintptr(coinitApartmentThreaded|coinitDisableOLE1DDE),
	)
	if hresultFailed(initializeResult) && uint32(initializeResult) != rpcEChangedMode {
		return fmt.Errorf("初始化 COM 失败: HRESULT 0x%08X", uint32(initializeResult))
	}
	if uint32(initializeResult) != rpcEChangedMode {
		defer procCoUninitialize.Call()
	}

	existingWindows := explorerWindowHandles()

	var folderItemIDList uintptr
	parseFolderResult, _, _ := procSHParseDisplayName.Call(
		uintptr(unsafe.Pointer(folderPathUTF16)),
		0,
		uintptr(unsafe.Pointer(&folderItemIDList)),
		0,
		0,
	)
	runtime.KeepAlive(folderPathUTF16)
	if hresultFailed(parseFolderResult) {
		return fmt.Errorf("解析文件夹 Shell 路径失败: HRESULT 0x%08X", uint32(parseFolderResult))
	}
	if folderItemIDList == 0 {
		return fmt.Errorf("解析文件夹 Shell 路径失败: 返回了空 ITEMIDLIST")
	}
	defer procCoTaskMemFree.Call(folderItemIDList)

	var fileItemIDList uintptr
	parseFileResult, _, _ := procSHParseDisplayName.Call(
		uintptr(unsafe.Pointer(filePathUTF16)),
		0,
		uintptr(unsafe.Pointer(&fileItemIDList)),
		0,
		0,
	)
	runtime.KeepAlive(filePathUTF16)
	if hresultFailed(parseFileResult) {
		return fmt.Errorf("解析文件 Shell 路径失败: HRESULT 0x%08X", uint32(parseFileResult))
	}
	if fileItemIDList == 0 {
		return fmt.Errorf("解析文件 Shell 路径失败: 返回了空 ITEMIDLIST")
	}
	defer procCoTaskMemFree.Call(fileItemIDList)

	childItemIDList, _, _ := procILFindLastID.Call(fileItemIDList)
	if childItemIDList == 0 {
		return fmt.Errorf("解析文件 Shell 子项目失败: 返回了空 ITEMIDLIST")
	}
	selectedItems := []uintptr{childItemIDList}

	// 明确传入父文件夹和相对文件 PIDL，确保首次创建 Explorer 窗口时也能完成文件选中。
	openResult, _, _ := procSHOpenFolderAndSelectItems.Call(
		folderItemIDList,
		1,
		uintptr(unsafe.Pointer(&selectedItems[0])),
		0,
	)
	runtime.KeepAlive(selectedItems)
	if hresultFailed(openResult) {
		return fmt.Errorf("打开文件夹并选中文件失败: HRESULT 0x%08X", uint32(openResult))
	}

	bringExplorerWindowToFront(existingWindows, folderPath)
	return nil
}

func hresultFailed(result uintptr) bool {
	return int32(uint32(result)) < 0
}

func explorerWindowHandles() map[uintptr]struct{} {
	result := make(map[uintptr]struct{})
	for _, window := range enumerateExplorerWindows() {
		result[window.handle] = struct{}{}
	}
	return result
}

func enumerateExplorerWindows() []explorerWindow {
	enumeration := explorerWindowEnumeration{}
	procEnumWindows.Call(enumExplorerWindowsCallback, uintptr(unsafe.Pointer(&enumeration)))
	runtime.KeepAlive(&enumeration)
	return enumeration.windows
}

func enumExplorerWindowCallback(windowHandle uintptr, parameter uintptr) uintptr {
	enumeration := (*explorerWindowEnumeration)(unsafe.Pointer(parameter))
	visible, _, _ := procIsWindowVisible.Call(windowHandle)
	if visible == 0 {
		return 1
	}

	classBuffer := make([]uint16, 256)
	classLength, _, _ := procGetClassNameW.Call(
		windowHandle,
		uintptr(unsafe.Pointer(&classBuffer[0])),
		uintptr(len(classBuffer)),
	)
	if classLength == 0 {
		return 1
	}
	className := windows.UTF16ToString(classBuffer[:classLength])
	if className != "CabinetWClass" && className != "ExploreWClass" {
		return 1
	}

	titleBuffer := make([]uint16, 512)
	titleLength, _, _ := procGetWindowTextW.Call(
		windowHandle,
		uintptr(unsafe.Pointer(&titleBuffer[0])),
		uintptr(len(titleBuffer)),
	)
	title := ""
	if titleLength > 0 {
		title = windows.UTF16ToString(titleBuffer[:titleLength])
	}
	enumeration.windows = append(enumeration.windows, explorerWindow{handle: windowHandle, title: title})
	return 1
}

func bringExplorerWindowToFront(existingWindows map[uintptr]struct{}, folderPath string) {
	deadline := time.Now().Add(2 * time.Second)
	folderName := strings.ToLower(filepath.Base(filepath.Clean(folderPath)))
	var fallbackWindow uintptr

	for time.Now().Before(deadline) {
		windowsList := enumerateExplorerWindows()
		for _, window := range windowsList {
			if _, existed := existingWindows[window.handle]; !existed {
				// 首次创建 Explorer 时，等待 Shell 完成目录导航和文件选中后再切到前台。
				time.Sleep(250 * time.Millisecond)
				activateWindow(window.handle)
				return
			}
		}
		for _, window := range windowsList {
			if folderName != "" && strings.Contains(strings.ToLower(window.title), folderName) {
				activateWindow(window.handle)
				return
			}
		}
		if len(windowsList) > 0 {
			// EnumWindows 按 Z 顺序返回；Shell 复用窗口时，首个 Explorer 通常就是刚更新的窗口。
			fallbackWindow = windowsList[0].handle
		}
		time.Sleep(75 * time.Millisecond)
	}

	if fallbackWindow != 0 {
		activateWindow(fallbackWindow)
	}
}

func activateWindow(windowHandle uintptr) {
	procShowWindowAsync.Call(windowHandle, swRestore)

	currentThread, _, _ := procGetCurrentThreadID.Call()
	targetThread, _, _ := procGetWindowThreadProcessID.Call(windowHandle, 0)
	foregroundWindow, _, _ := procGetForegroundWindow.Call()
	foregroundThread := uintptr(0)
	if foregroundWindow != 0 {
		foregroundThread, _, _ = procGetWindowThreadProcessID.Call(foregroundWindow, 0)
	}

	if foregroundThread != 0 && foregroundThread != currentThread {
		procAttachThreadInput.Call(currentThread, foregroundThread, 1)
		defer procAttachThreadInput.Call(currentThread, foregroundThread, 0)
	}
	if targetThread != 0 && targetThread != currentThread && targetThread != foregroundThread {
		procAttachThreadInput.Call(currentThread, targetThread, 1)
		defer procAttachThreadInput.Call(currentThread, targetThread, 0)
	}

	procBringWindowToTop.Call(windowHandle)
	procSetForegroundWindow.Call(windowHandle)

	// 调整到普通窗口 Z 序的最前面，不设置永久置顶属性。
	flags := uintptr(swpNoSize | swpNoMove | swpShowWindow)
	procSetWindowPos.Call(windowHandle, 0, 0, 0, 0, 0, flags) // HWND_TOP = 0
	procSetForegroundWindow.Call(windowHandle)
}
