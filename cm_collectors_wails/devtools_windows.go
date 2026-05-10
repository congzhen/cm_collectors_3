package main

import (
	"syscall"
	"time"
)

const (
	vkControl       = 0x11
	vkShift         = 0x10
	vkF12           = 0x7B
	keyeventfKeyUp  = 0x0002
	keyPressDelayMS = 10
)

var procKeybdEvent = syscall.NewLazyDLL("user32.dll").NewProc("keybd_event")

// OpenDevToolsShortcut asks Wails/WebView2 to open DevTools.
func (a *App) OpenDevToolsShortcut() error {
	keyDown(vkControl)
	keyDown(vkShift)
	keyDown(vkF12)
	time.Sleep(keyPressDelayMS * time.Millisecond)
	keyUp(vkF12)
	keyUp(vkShift)
	keyUp(vkControl)
	return nil
}

func keyDown(vk byte) {
	procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
}

func keyUp(vk byte) {
	procKeybdEvent.Call(uintptr(vk), 0, keyeventfKeyUp, 0)
}
