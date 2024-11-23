//go:build js

package procutil

import "os"

func WaitForSigterm() os.Signal {
	select {}
	return nil
}

func SelfSIGHUP() {

}

func NewSighupChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	return ch
}
