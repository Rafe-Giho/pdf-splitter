//go:build windows

package platform

import (
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"gioui.org/app"
)

const (
	wmDropFiles = 0x0233
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc   = user32.NewProc("CallWindowProcW")
	procDefWindowProc    = user32.NewProc("DefWindowProcW")
	procDragAcceptFiles  = shell32.NewProc("DragAcceptFiles")
	procDragQueryFile    = shell32.NewProc("DragQueryFileW")
	procDragFinish       = shell32.NewProc("DragFinish")
	dropProc             = syscall.NewCallback(dropWindowProc)
	dropTargets          sync.Map
)

type dropTarget struct {
	oldProc uintptr
	onFile  func(string)
}

func AttachFileDrop(viewEvent any, onFile func(string)) error {
	winEvent, ok := viewEvent.(app.Win32ViewEvent)
	if !ok || winEvent.HWND == 0 || onFile == nil {
		return nil
	}

	if current, ok := dropTargets.Load(winEvent.HWND); ok {
		current.(*dropTarget).onFile = onFile
		return nil
	}

	oldProc, _, _ := procSetWindowLongPtr.Call(
		winEvent.HWND,
		^uintptr(3),
		dropProc,
	)
	if oldProc == 0 {
		return nil
	}

	procDragAcceptFiles.Call(winEvent.HWND, 1)
	dropTargets.Store(winEvent.HWND, &dropTarget{
		oldProc: oldProc,
		onFile:  onFile,
	})

	return nil
}

func dropWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if target, ok := dropTargets.Load(hwnd); ok {
		state := target.(*dropTarget)
		if msg == wmDropFiles {
			handleDrop(wParam, state.onFile)
			return 0
		}

		if state.oldProc == 0 {
			ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
			return ret
		}

		ret, _, _ := procCallWindowProc.Call(state.oldProc, hwnd, uintptr(msg), wParam, lParam)
		return ret
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func handleDrop(hDrop uintptr, onFile func(string)) {
	defer procDragFinish.Call(hDrop)

	count, _, _ := procDragQueryFile.Call(hDrop, 0xFFFFFFFF, 0, 0)
	if count == 0 {
		return
	}

	for i := uintptr(0); i < count; i++ {
		length, _, _ := procDragQueryFile.Call(hDrop, i, 0, 0)
		if length == 0 {
			continue
		}

		buf := make([]uint16, length+1)
		procDragQueryFile.Call(hDrop, i, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		path := filepath.Clean(syscall.UTF16ToString(buf))
		if strings.EqualFold(filepath.Ext(path), ".pdf") {
			go onFile(path)
			return
		}
	}
}
