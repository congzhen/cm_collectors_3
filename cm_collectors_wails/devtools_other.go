//go:build !windows

package main

// OpenDevToolsShortcut is implemented for Windows, where this Wails shell runs.
func (a *App) OpenDevToolsShortcut() error {
	return nil
}
