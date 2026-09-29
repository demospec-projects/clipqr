package main

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	whKeyboardLL = 13
	wmKeyDown    = 0x0100
	wmSysKeyDown = 0x0104
	wmQuit       = 0x0012
	vkSnapshot   = 0x2C
)

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procSetWindowsHookEx  = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx    = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHook = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage        = user32.NewProc("GetMessageW")
	procPostThreadMessage = user32.NewProc("PostThreadMessageW")

	// One callback for the life of the program: Windows limits how many a
	// Go program may create.
	hookOnce     sync.Once
	hookCallback uintptr
	hookPress    func()
)

type kbdLLHook struct {
	VkCode, ScanCode, Flags, Time uint32
	ExtraInfo                     uintptr
}

// printKey swallows Print Screen with a low-level keyboard hook, before
// Windows hands it to the Snipping Tool.
type printKey struct {
	mu       sync.Mutex
	threadID uint32
}

func newPrintKey(string) *printKey { return &printKey{} }

func printKeyAvailable() bool { return true }

func (p *printKey) start(onPress func()) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.threadID != 0 {
		return nil
	}
	hookPress = onPress
	hookOnce.Do(func() {
		hookCallback = windows.NewCallback(func(code, wparam, lparam uintptr) uintptr {
			if int32(code) >= 0 {
				// lparam points to memory Windows owns, not the Go heap.
				key := *(**kbdLLHook)(unsafe.Pointer(&lparam))
				if key.VkCode == vkSnapshot {
					if (wparam == wmKeyDown || wparam == wmSysKeyDown) && hookPress != nil {
						go hookPress()
					}
					return 1 // swallowed: neither Windows nor the Snipping Tool sees it
				}
			}
			r, _, _ := procCallNextHookEx.Call(0, code, wparam, lparam)
			return r
		})
	})
	ready := make(chan error)
	go func() {
		// The hook lives on this thread, which must keep reading its messages.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var module windows.Handle
		windows.GetModuleHandleEx(0, nil, &module)
		hook, _, err := procSetWindowsHookEx.Call(whKeyboardLL, hookCallback, uintptr(module), 0)
		if hook == 0 {
			ready <- errors.New("Impossible de prendre la touche Impr. écran : " + err.Error())
			return
		}
		p.threadID = windows.GetCurrentThreadId()
		ready <- nil
		var msg [64]byte
		for {
			r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
			if int32(r) <= 0 {
				break
			}
		}
		procUnhookWindowsHook.Call(hook)
	}()
	return <-ready
}

func (p *printKey) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.threadID != 0 {
		procPostThreadMessage.Call(uintptr(p.threadID), wmQuit, 0, 0)
		p.threadID = 0
	}
	hookPress = nil
}

func (p *printKey) restoreDesktopBindings() {}
