package fs

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs/fsproxy"
)

func fadviseSequentialRead(_ *fsproxy.ProxyFile, _ bool) error {
	// TODO: implement this properly
	return nil
}
