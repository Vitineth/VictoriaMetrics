//go:build js

package httpserver

import (
	wasmhttp "github.com/nlepage/go-wasm-http-server"
)

func serve(addr string, useProxyProtocol bool, rh RequestHandler, idx int) {
	wasmhttp.Serve(gzipHandler(&server{}, rh))
}
