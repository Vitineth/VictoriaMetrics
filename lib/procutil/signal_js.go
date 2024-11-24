//go:build js

package procutil

import (
	"log/slog"
	"os"
	"syscall"
	"syscall/js"
)

func WaitForSigterm() os.Signal {
	slog.Info("Waiting for sigterm using global func")
	ch := make(chan os.Signal, 1)
	js.Global().Set("terminateVictoriaMetricsInstance", js.FuncOf(func(this js.Value, args []js.Value) any {
		ch <- syscall.SIGTERM
		return nil
	}))
	slog.Info("property should be set")

	for {
		sig := <-ch
		return sig
	}
}

func SelfSIGHUP() {

}

func NewSighupChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	return ch
}
