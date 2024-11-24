//go:build !js

package httpserver

import (
	"context"
	"crypto/tls"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fasttime"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/netutil"
	"github.com/valyala/fastrand"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

type serverWrapper struct {
	shutdownDelayDeadline *atomic.Int64
	s                     *http.Server
}

func (s serverWrapper) Stop(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

func (s serverWrapper) ShutdownDelayDeadline() *atomic.Int64 {
	return s.shutdownDelayDeadline
}

func serve(addr string, useProxyProtocol bool, rh RequestHandler, idx int) {
	scheme := "http"
	if tlsEnable.GetOptionalArg(idx) {
		scheme = "https"
	}
	var tlsConfig *tls.Config
	if tlsEnable.GetOptionalArg(idx) {
		certFile := tlsCertFile.GetOptionalArg(idx)
		keyFile := tlsKeyFile.GetOptionalArg(idx)
		minVersion := tlsMinVersion.GetOptionalArg(idx)
		tc, err := netutil.GetServerTLSConfig(certFile, keyFile, minVersion, *tlsCipherSuites)
		if err != nil {
			logger.Fatalf("cannot load TLS cert from -tlsCertFile=%q, -tlsKeyFile=%q, -tlsMinVersion=%q, -tlsCipherSuites=%q: %s", certFile, keyFile, minVersion, *tlsCipherSuites, err)
		}
		tlsConfig = tc
	}
	ln, err := netutil.NewTCPListener(scheme, addr, useProxyProtocol, tlsConfig)
	if err != nil {
		logger.Fatalf("cannot start http server at %s: %s", addr, err)
	} else {
		logger.Infof("started server at %s://%s/", scheme, ln.Addr())
		logger.Infof("pprof handlers are exposed at %s://%s/debug/pprof/", scheme, ln.Addr())
	}
	serveWithListener(addr, ln, rh)
}

func serveWithListener(addr string, ln net.Listener, rh RequestHandler) {
	var s serverWrapper
	s.s = &http.Server{
		Handler: gzipHandler(&s, rh),

		// Disable http/2, since it doesn't give any advantages for VictoriaMetrics services.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),

		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       *idleConnTimeout,

		// Do not set ReadTimeout and WriteTimeout here,
		// since these timeouts must be controlled by request handlers.

		ErrorLog: logger.StdErrorLogger(),
	}
	if *connTimeout > 0 {
		s.s.ConnContext = func(ctx context.Context, _ net.Conn) context.Context {
			timeoutSec := connTimeout.Seconds()
			// Add a jitter for connection timeout in order to prevent Thundering herd problem
			// when all the connections are established at the same time.
			// See https://en.wikipedia.org/wiki/Thundering_herd_problem
			jitterSec := fastrand.Uint32n(uint32(timeoutSec / 10))
			deadline := fasttime.UnixTimestamp() + uint64(timeoutSec) + uint64(jitterSec)
			return context.WithValue(ctx, connDeadlineTimeKey, &deadline)
		}
	}

	serversLock.Lock()
	servers[addr] = &s
	serversLock.Unlock()
	if err := s.s.Serve(ln); err != nil {
		if err == http.ErrServerClosed {
			// The server gracefully closed.
			return
		}
		logger.Panicf("FATAL: cannot serve http at %s: %s", addr, err)
	}
}
