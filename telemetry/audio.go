package telemetry

import (
	"sync"
	"syscall"
	"time"
)

var (
	kernel32DLL = syscall.NewLazyDLL("kernel32.dll")
	procBeep    = kernel32DLL.NewProc("Beep")

	audioMu sync.Mutex
)

// PlayStreamConnected spielt einen aufsteigenden Signalton ab
func PlayStreamConnected() {
	go func() {
		audioMu.Lock()
		defer audioMu.Unlock()

		_, _, _ = procBeep.Call(784, 90) // G5
		time.Sleep(30 * time.Millisecond)
		_, _, _ = procBeep.Call(1046, 130) // C6
	}()
}

// PlayStreamDisconnected spielt einen absteigenden Warn-Signalton ab
func PlayStreamDisconnected() {
	go func() {
		audioMu.Lock()
		defer audioMu.Unlock()

		_, _, _ = procBeep.Call(659, 120) // E5
		time.Sleep(40 * time.Millisecond)
		_, _, _ = procBeep.Call(440, 220) // A4
	}()
}
