//go:build js

package httpserver

import (
	"context"
	wasmhttp "github.com/nlepage/go-wasm-http-server"
	"sync/atomic"
)

var a atomic.Int64

type wrapper struct {
	terminator func()
}

func (w wrapper) Stop(ctx context.Context) error {
	w.terminator()
	return nil
}

func (w wrapper) ShutdownDelayDeadline() *atomic.Int64 {
	return &a
}

func serve(addr string, useProxyProtocol bool, rh RequestHandler, idx int) {
	var w wrapper
	w.terminator = wasmhttp.Serve(gzipHandler(w, rh))
	servers[addr] = w
}
