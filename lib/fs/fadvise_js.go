//go:build js

package fs

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs/fsproxy"
)

func fadviseSequentialRead(f *fsproxy.ProxyFile, prefetch bool) error {
	return nil
}
